package main

import (
	"fmt"
	"net"
	"testing"
)

func TestSendMsg(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		t.Error("conn err")
	}
	s := "*1\r\n$7\r\nCOMMAND\r\n"
	n, _ := conn.Write([]byte(s))
	fmt.Println("send", n)

	buf := make([]byte, 1024)
	n, _ = conn.Read(buf)
	fmt.Println("recv: ["+string(buf[:n])+"] n:", n)
}
