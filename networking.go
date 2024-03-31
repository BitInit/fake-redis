package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"syscall"

	"github.com/BitInit/fake-redis/ae"
	"github.com/BitInit/fake-redis/anet"
	"github.com/BitInit/fake-redis/sds"
	"github.com/BitInit/fake-redis/util"
)

// client request types
const protoNone = 0
const protoReqInline = 1
const protoReqMultibulk = 2

const protoReplyChunkBytes = (16 * 1024) // 16k output buffer

const PROTO_INLINE_MAX_SIZE = (1024 * 64)

type client struct {
	id uint64
	fd int
	db int

	reqtype  int
	qb_pos   int
	querybuf *sds.Sds

	multibulklen int // Number of multi bulk arguments left to read.
	bulklen      int // Length of bulk argument in multi bulk request.
	argc         int
	argv         []*robj
	cmd          *redisCommand
	lastcmd      *redisCommand

	// repsonse buffer
	buf     [protoReplyChunkBytes]byte
	bufpos  int
	sentlen int // Amount of bytes already sent in the current buffer
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
	if err := anet.EnableTcpNoDelay(cfd); err != nil {
		log.Println("enable tcp no delay error", err)
	}

	c := &client{}
	if err := server.el.CreateFileEvent(cfd, ae.AE_READABLE, readQueryFromClient, c); err != nil {
		syscall.Close(cfd)
		return nil
	}

	c.id = server.next_client_id.Add(1)
	c.fd = cfd
	selectDb(c, 0)
	c.reqtype = 0
	c.qb_pos = 0
	c.bulklen = -1
	c.querybuf = sds.Empty()
	c.argc = 0
	c.cmd = nil
	c.lastcmd = nil
	return c
}

func freeClient(c *client) {
	server.el.DeleteFileEvent(c.fd, ae.AE_READABLE)
	server.el.DeleteFileEvent(c.fd, ae.AE_WRITABLE)

	syscall.Close(c.fd)
}

func resetClient(c *client) {
	c.argc = 0
	c.argv = make([]*robj, 0)

	c.reqtype = protoNone
	c.multibulklen = 0
	c.bulklen = -1
}

func processMultibulkBuffer(c *client) bool {
	if c.multibulklen == 0 {
		buf := c.querybuf.Buf
		idx := util.SliceIndexByte(buf, '\r', c.qb_pos)
		if idx == -1 {
			if c.querybuf.Len()-c.qb_pos > PROTO_INLINE_MAX_SIZE {
				addReplyError(c, "Protocol error: too big mbulk count string")
			}
			return false
		}
		ll, err := strconv.Atoi(string(buf[c.qb_pos+1 : idx]))
		if err != nil || ll > 1024*1024 {
			addReplyError(c, "Protocol error: invalid multibulk length")
			return false
		}
		if ll <= 0 {
			return true
		}
		c.qb_pos += idx + 2
		c.multibulklen = ll
		c.argv = make([]*robj, ll)
	}

	for c.multibulklen > 0 {
		if c.bulklen == -1 {
			idx := util.SliceIndexByte(c.querybuf.Buf, '\r', c.qb_pos)
			if idx == -1 {
				break
			}
			if c.querybuf.Buf[c.qb_pos] != '$' {
				addReplyErrorFormat(c, "Protocol error: expected '$', got '%c'",
					c.querybuf.Buf[c.qb_pos])
				return false
			}
			ll, err := strconv.Atoi(string(c.querybuf.Buf[c.qb_pos+1 : idx]))
			if err != nil {
				addReplyError(c, "Protocol error: invalid bulk length")
				return false
			}
			c.qb_pos = idx + 2
			c.bulklen = ll
		}

		if c.querybuf.Len()-c.qb_pos < c.bulklen+2 {
			break
		} else {
			c.argv[c.argc] = createStringObject(c.querybuf.Buf[c.qb_pos : c.qb_pos+c.bulklen])
			c.argc++
			c.qb_pos += c.bulklen + 2
		}
		c.bulklen = -1
		c.multibulklen--
	}

	return c.multibulklen == 0
}

func processInputBuffer(c *client) {
	server.current_client = c
	for c.qb_pos < c.querybuf.Len() {
		if c.reqtype == protoNone {
			if c.querybuf.Buf[0] == '*' {
				c.reqtype = protoReqMultibulk
			} else {
				c.reqtype = protoReqInline
			}
		}

		if c.reqtype == protoReqMultibulk {
			if !processMultibulkBuffer(c) {
				break
			}
		}

		if c.argc == 0 {
			resetClient(c)
		} else {
			if processCommand(c) {
				resetClient(c)
			}
			if server.current_client == nil {
				break
			}
		}
	}

	if server.current_client != nil && c.qb_pos != 0 {
		c.querybuf = c.querybuf.Range(c.qb_pos, -1)
		c.qb_pos = 0
	}
	server.current_client = nil
}

func processInputBufferAndReplicate(c *client) {
	processInputBuffer(c)
}

func readQueryFromClient(el *ae.EventLoop, s int, privdata interface{}, mask int) {
	c := privdata.(*client)
	readlen := 10

	qblen := c.querybuf.Len()
	c.querybuf = c.querybuf.MakeRoomFor(readlen)
	nread, err := syscall.Read(c.fd, c.querybuf.Buf[qblen:])
	if err == io.EOF || nread == 0 {
		log.Println("client closed connection")
		freeClient(c)
		return
	} else if err != nil {
		log.Println("reading from client error", err)
		freeClient(c)
		return
	}
	c.querybuf.IncLen(nread)

	processInputBufferAndReplicate(c)
}

// ============= Client reponse data ==============
func clientInstallWriteHandler(c *client) {
	server.clients_pending_write.AddNodeHead(c)
}

func prepareClientToWrite(c *client) bool {
	if !clientHasPendingReplies(c) {
		clientInstallWriteHandler(c)
	}
	return true
}

func addReply(c *client, obj *robj) {
	if !prepareClientToWrite(c) {
		return
	}

	if sdsEncodedObject(obj) {
		bs := obj.ptr.([]byte)
		if !_addReplyToBuffer(c, bs) {
			log.Panicln("send data error")
		}
	}
}

func addReplyString(c *client, s string) {
	if !prepareClientToWrite(c) {
		return
	}
	if !_addReplyToBuffer(c, []byte(s)) {
		log.Panicln("send data error")
	}
}

func addReplyError(c *client, s string) {
	if len(s) == 0 || s[0] != '-' {
		addReplyString(c, "-ERR ")
	}
	addReplyString(c, s)
	addReplyString(c, "\r\n")
}

func addReplyErrorFormat(c *client, s string, args ...any) {
	msg := fmt.Sprintf(s, args...)
	var resp strings.Builder
	for i := 0; i < len(msg); i++ {
		resp.WriteByte(msg[i])
	}
	addReplyError(c, resp.String())
}

func _addReplyToBuffer(c *client, b []byte) bool {
	avial := len(c.buf) - c.bufpos
	if len(b) > avial {
		return false
	}
	copy(c.buf[c.bufpos:], b)
	c.bufpos += len(b)
	return true
}

func clientHasPendingReplies(c *client) bool {
	return c.bufpos > 0
}

func writeToClient(fd int, c *client, handler_installed bool) bool {
	for clientHasPendingReplies(c) {
		if c.bufpos > 0 {
			nwritten, err := syscall.Write(fd, c.buf[c.sentlen:c.bufpos])
			if err != nil {
				break
			}
			c.sentlen += nwritten

			if c.sentlen == c.bufpos {
				c.bufpos = 0
				c.sentlen = 0
			}
		}
	}
	return false
}

func handleClientsWithPendingWrites() int {
	list := server.clients_pending_write
	processed := list.Length()

	li := list.Iterator(true)
	for ln := li.Next(); ln != nil; ln = li.Next() {
		c := ln.NodeValue().(*client)
		list.DelNode(ln)

		if !writeToClient(c.fd, c, false) {
			continue
		}
		// TODO
	}
	return int(processed)
}
