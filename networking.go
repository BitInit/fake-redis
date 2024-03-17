package main

import (
	"fmt"
	"io"
	"log"
	"syscall"

	"github.com/BitInit/fake-redis/ae"
	"github.com/BitInit/fake-redis/anet"
)

type client struct {
	fd int
	db int

	qb_pos   int
	querybuf []byte
}

func acceptTcpHandler(el *ae.EventLoop, s int, privdata interface{}, mask int) {
	cfd, sa, err := anet.TcpAccept(s)
	if err != nil {
		log.Println("Accepting client connection error", err)
		return
	}

	sav4 := sa.(*syscall.SockaddrInet4)
	log.Printf("accepted %d.%d.%d.%d:%d", sav4.Addr[0], sav4.Addr[1], sav4.Addr[2], sav4.Addr[3], sav4.Port)
	createClient(cfd)
}

func createClient(cfd int) *client {
	anet.NonBlock(cfd)
	anet.EnableTcpNoDelay(cfd)

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

func freeClient(c *client) {
	server.el.DeleteFileEvent(c.fd, ae.AE_READABLE)
	server.el.DeleteFileEvent(c.fd, ae.AE_WRITABLE)

	syscall.Close(c.fd)
}

func processInputBuffer(c *client) {
	s := string(c.querybuf)
	fmt.Println(s[:c.qb_pos])
}

func processInputBufferAndReplicate(c *client) {
	processInputBuffer(c)
}

func readQueryFromClient(el *ae.EventLoop, s int, privdata interface{}, mask int) {
	c := privdata.(*client)
	readlen := PROTO_IOBUF_LEN

	c.querybuf = make([]byte, readlen)
	n, err := syscall.Read(c.fd, c.querybuf)
	if err == io.EOF || n == 0 {
		log.Println("client closed connection")
		freeClient(c)
		return
	} else if err != nil {
		log.Println("reading from client error", err)
		freeClient(c)
		return
	}

	c.qb_pos = n
	processInputBufferAndReplicate(c)
}
