package ae

import (
	"syscall"
	"time"
)

type apiState struct {
	epfd   int
	events []syscall.EpollEvent
}

func apiCreate(eventLoop *EventLoop) error {
	fd, err := syscall.EpollCreate(1024)
	if err != nil {
		return err
	}

	state := &apiState{
		epfd:   fd,
		events: make([]syscall.EpollEvent, eventLoop.setsize),
	}
	eventLoop.apidata = state
	return nil
}

func apiPoll(el *EventLoop, t *time.Duration) int {
	state := el.apidata.(*apiState)

	msec := -1
	if t != nil {
		msec = int(t.Milliseconds())
	}
	n, err := syscall.EpollWait(state.epfd, state.events, msec)
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

func apiAddEvent(el *EventLoop, fd int, mask int) error {
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
		return err
	}
	return nil
}

func apiDelEvent(el *EventLoop, fd int, delmask int) {
	state := el.apidata.(*apiState)
	mask := el.events[fd].mask & (^delmask)

	ee := syscall.EpollEvent{}
	if mask&AE_READABLE != 0 {
		ee.Events |= syscall.EPOLLIN
	}
	if mask&AE_WRITABLE != 0 {
		ee.Events |= syscall.EPOLLOUT
	}
	ee.Fd = int32(fd)
	if mask != AE_NONE {
		syscall.EpollCtl(state.epfd, syscall.EPOLL_CTL_MOD, fd, &ee)
	} else {
		syscall.EpollCtl(state.epfd, syscall.EPOLL_CTL_DEL, fd, &ee)
	}
}
