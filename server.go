package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BitInit/fake-redis/ae"
	"github.com/BitInit/fake-redis/anet"
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

	dbnum int
}

var server redisServer

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
