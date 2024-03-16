package ae

import (
	"log"
	"syscall"
)

type apiState struct {
	epfd   int
	events []syscall.EpollEvent
}

func apiCreate(eventLoop *EventLoop) bool {
	fd, err := syscall.EpollCreate(1024)
	if err != nil {
		log.Println("epoll apiCreate failed")
		return false
	}

	state := apiState{
		epfd:   fd,
		events: make([]syscall.EpollEvent, eventLoop.setsize),
	}
	eventLoop.apidata = &state
	return true
}

func apiPoll(el *EventLoop) int {
	state := el.apidata.(*apiState)
	n, err := syscall.EpollWait(state.epfd, state.events, -1)
	if err != nil {
		return 0
	}
	for i := 0; i < n; i++ {
		mask := 0
		e := state.events[i]

		if e.Events&syscall.EPOLLIN != 0 {
			mask |= AE_READABLE
		}
		if e.Events&syscall.EPOLLOUT != 0 {
			mask |= AE_WRITABLE
		}
		if e.Events&syscall.EPOLLERR != 0 {
			mask |= AE_WRITABLE
		}
		if e.Events&syscall.EPOLLHUP != 0 {
			mask |= AE_WRITABLE
		}
		el.fired[i].fd = int(e.Fd)
		el.fired[i].mask = int(e.Events)
	}
	return n
}

func apiAddEvent(el *EventLoop, fd int, mask int) bool {
	state := el.apidata.(*apiState)
	op := syscall.EPOLL_CTL_MOD
	if el.events[fd].mask == AE_NONE {
		op = syscall.EPOLL_CTL_ADD
	}

	ee := syscall.EpollEvent{}
	mask |= el.events[fd].mask
	if mask&AE_READABLE != 0 {
		ee.Events |= syscall.EPOLLIN
	}
	if mask&AE_WRITABLE != 0 {
		ee.Events |= syscall.EPOLLOUT
	}
	ee.Fd = int32(fd)
	if err := syscall.EpollCtl(state.epfd, op, fd, &ee); err != nil {
		return false
	}
	return true
}
