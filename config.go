package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
)

const CONFIG_DEFAULT_HZ = 10 // Time interrupt calls/sec.
const CONFIG_MIN_HZ = 1
const CONFIG_BINDADDR_MAX = 16
const CONFIG_DEFAULT_SERVER_PORT = 6379
const CONFIG_DEFAULT_MAX_CLIENTS = 10000
const CONFIG_DEFAULT_TCP_BACKLOG = 511

const PROTO_IOBUF_LEN = (1024 * 16) /* Generic I/O buffer size */

func initServerConfig() {
	server.maxclients = CONFIG_DEFAULT_MAX_CLIENTS
	server.port = CONFIG_DEFAULT_SERVER_PORT
	server.tcp_backlog = CONFIG_DEFAULT_TCP_BACKLOG
	server.dbnum = 16
}

func loadServerConfig(confFile string, options string) {
	var config string
	if len(confFile) > 0 {
		content, err := os.ReadFile(confFile)
		if err != nil {
			log.Fatalln("Fatal error, can't read config file ", confFile)
		}
		config += string(content)
	}
	if len(options) > 0 {
		config += "\n"
		config += options
	}
	loadServerConfigFromString(config)
}

func loadServerConfigFromString(config string) {
	lines := strings.Split(config, "\n")
	var err error
	for _, line := range lines {
		line = strings.Trim(line, " \t\r\n")
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		argv := strings.Split(line, " ")
		argc := len(argv)
		if argc == 0 {
			continue
		}

		argv[0] = strings.ToLower(argv[0])
		if argv[0] == "port" && argc == 2 {
			if port, _err := strconv.Atoi(argv[1]); port < 0 || port > 65535 || _err != nil {
				err = errors.New("invalid port")
				goto loaderr
			} else {
				server.port = port
			}
		}
	}
	return

loaderr:
	os.Stderr.WriteString("\n*** FATAL CONFIG FILE ERROR ***\n")
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}
