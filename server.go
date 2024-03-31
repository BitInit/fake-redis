package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/BitInit/fake-redis/adlist"
	"github.com/BitInit/fake-redis/ae"
	"github.com/BitInit/fake-redis/anet"
	"github.com/BitInit/fake-redis/dict"
	"github.com/BitInit/fake-redis/util"
)

type redisServer struct {
	configFile string
	port       int
	maxclients int

	ipfd        []int
	tcp_backlog int
	// bindAddr    []string
	// setsize     int
	el *ae.EventLoop

	commands *dict.Dict
	dbnum    int
	db       []*redisDb

	next_client_id atomic.Uint64
	current_client *client

	clients_pending_write *adlist.List
}

var server redisServer

type redisCommandProc func(c *client)
type redisCommand struct {
	name string
	proc redisCommandProc
}

var redisCommandTalbe []*redisCommand = []*redisCommand{
	{name: "get", proc: getCommand},
	{name: "set", proc: setCommand},
	{name: "setnx", proc: setnxCommand},
	{name: "del", proc: delCommand},
	{name: "command", proc: commandCommand},
}

func commandCommand(c *client) {
	addReplyString(c, "+OK\r\n")
}

// ====================== share data ====================
type sharedObjectsStruct struct {
	ok *robj // +ok\r\n
}

var shared sharedObjectsStruct

func createSharedObjects() {
	shared.ok = createStringObject([]byte("+OK\r\n"))
}

func main() {
	initServerConfig()
	loadConfig()

	initServer()

	ae.AeMain(server.el)
}

func loadConfig() {
	args := os.Args
	argc := len(args)
	if argc >= 2 {
		if args[1] == "-v" || args[1] == "--version" {
			version()
		}
		if args[1] == "-h" || args[1] == "--help" {
			usage()
		}

		j := 1
		var configFile string
		if !strings.HasPrefix(args[j], "-") {
			configFile = args[j]
			server.configFile = util.GetAbolutePath(configFile)
			j++
		}

		var options string
		for j != argc {
			if strings.HasPrefix(args[j], "--") {
				if len(options) == 0 {
					options += "\n"
				}
				options += args[j][2:]
				options += " "
			} else {
				options += args[j]
				options += " "
			}
			j++
		}
		loadServerConfig(configFile, options)
	}
}

func version() {
	fmt.Println("Redis server v = 5.0.8")
	os.Exit(0)
}

func usage() {
	fmt.Println("Usage:")
	os.Exit(0)
}

func initServer() {
	el, err := ae.CreateEventLoop(server.maxclients)
	if err != nil {
		log.Fatalln("failed creating the event loop.", err)
	}
	server.el = el
	listenToPort()
	server.el.CreateFileEvent(server.ipfd[0], ae.AE_READABLE, acceptTcpHandler, nil)
	server.el.SetBeforeSleepProc(beforeSleepProc)
	server.clients_pending_write = adlist.Create()

	createSharedObjects()

	for i := 0; i < server.dbnum; i++ {
		rdb := &redisDb{
			id:      i,
			dict:    dict.Create(dbDictType, nil),
			expires: dict.Create(keyptrDictType, nil),
		}
		server.db = append(server.db, rdb)
	}

	server.commands = dict.Create(commandTableDictType, nil)
	populateCommandTable()
}

func populateCommandTable() {
	for _, cmd := range redisCommandTalbe {
		server.commands.Add([]byte(cmd.name), cmd)
	}
}

func beforeSleepProc() {
	// send data to client
	handleClientsWithPendingWrites()
}

func listenToPort() {
	if server.port == 0 {
		log.Fatalln("port is 0")
	}
	ipfd, err := anet.TcpServer(server.port, "", server.tcp_backlog)
	if err != nil {
		log.Fatalln("tcpServer failed", err)
	}
	server.ipfd = append(server.ipfd, ipfd)
	anet.NonBlock(ipfd)
}

// ============== Commands lookup and execution ==============
func lookupCommand(bname []byte) *redisCommand {
	cmd := server.commands.GetVal(bname)
	if cmd == nil {
		return nil
	}
	return cmd.(*redisCommand)
}

func processCommand(c *client) bool {
	argv0 := c.argv[0].ptr.([]byte)
	if strings.EqualFold(string(argv0), "quit") {
		addReply(c, shared.ok)
		// TODO return C_ERR of redis source, why?
		return true
	}

	cmd := lookupCommand(argv0)
	if cmd == nil {
		var args strings.Builder
		for i := 1; i < c.argc && args.Len() < 128; i++ {
			args.WriteString(string(c.argv[i].ptr.([]byte)))
		}
		addReplyErrorFormat(c,
			"unkown command `%s`, with args beginning with: %s",
			string(argv0), args.String())
		return true
	}
	c.cmd = cmd
	c.lastcmd = cmd

	cmd.proc(c)
	return true
}

// =======================================
// Utility functions
// =======================================

func dictBytesHash(key interface{}) uint64 {
	buf := key.([]byte)

	var hash uint64 = 5381
	for i := 0; i < len(buf); i++ {
		hash = ((hash << 5) + hash) + uint64(buf[i])
	}
	return hash
}

func dictBytesKeyCompare(privdata, key1, key2 interface{}) int {
	buf1 := key1.([]byte)
	buf2 := key2.([]byte)

	if len(buf1) != len(buf2) {
		return 0
	}
	return bytes.Compare(buf1, buf2)
}

func dictBytesStrCaseCompare(privdata, key1, key2 interface{}) int {
	s1 := strings.ToLower(string(key1.([]byte)))
	s2 := strings.ToLower(string(key2.([]byte)))

	return strings.Compare(s1, s2)
}

// Db->dict, keys are []byte, vals are Redis objects
var dbDictType *dict.DictType = &dict.DictType{
	HashFunction: dictBytesHash,
	KeyDup:       nil,
	ValDup:       nil,
	KeyCompare:   dictBytesKeyCompare,
}

var keyptrDictType *dict.DictType = &dict.DictType{
	HashFunction: dictBytesHash,
	KeyDup:       nil,
	ValDup:       nil,
	KeyCompare:   dictBytesKeyCompare,
}

var commandTableDictType *dict.DictType = &dict.DictType{
	HashFunction: dictBytesHash,
	KeyDup:       nil,
	ValDup:       nil,
	KeyCompare:   dictBytesStrCaseCompare,
}
