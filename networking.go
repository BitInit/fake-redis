package main

import (
	"fmt"
	"log"
	"syscall"

	"github.com/BitInit/fake-redis/ae"
)

type client struct {
	fd int
	db int

	querybuf []byte
}

func tcpAcceptHandler(el *ae.EventLoop, s int, privdata interface{}, mask int) {
	cfd, _, err := syscall.Accept(s)
	if err != nil {
		log.Println("Accepting client connection error")
		return
	}
	createClient(cfd)
}

func readQueryFromClient(el *ae.EventLoop, s int, privdata interface{}, mask int) {
	c := privdata.(*client)
	readlen := PROTO_IOBUF_LEN

	c.querybuf = make([]byte, readlen)
	if _, err := syscall.Read(c.fd, c.querybuf); err != nil {
	}
	fmt.Println(string(c.querybuf))
}

func createClient(cfd int) *client {
	c := &client{
		fd: cfd,
	}
	if err := server.el.CreateFileEvent(cfd, ae.AE_READABLE, readQueryFromClient, c); err != nil {
		syscall.Close(cfd)
		return nil
	}

	selectDb(c, 0)
	return c
}
