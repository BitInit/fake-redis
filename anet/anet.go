package anet

import (
	"syscall"
)

func TcpServer(port int, bindaddr string, backlog int) (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return -1, nil
	}
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return -1, err
	}

	sa := &syscall.SockaddrInet4{
		Port: port,      // Listen on this port number
		Addr: [4]byte{}, // Listen to all IPs
	}
	err = syscall.Bind(fd, sa)
	if err != nil {
		return -1, err
	}
	err = syscall.Listen(fd, backlog)
	if err != nil {
		return -1, err
	}
	return fd, nil
}

func TcpAccept(fd int) (int, syscall.Sockaddr, error) {
	cfd, sa, err := syscall.Accept(fd)
	if err != nil {
		return -1, nil, err
	}
	return cfd, sa, nil
}

func NonBlock(fd int) error {
	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}
	return nil
}

func EnableTcpNoDelay(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOCK_STREAM, syscall.TCP_NODELAY, 1)
}
