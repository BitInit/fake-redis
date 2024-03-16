package ae

import (
	"errors"
	"time"
)

const AE_NONE = 0
const AE_READABLE = 1
const AE_WRITABLE = 2
const AE_BARRIER = 4

type EventLoop struct {
	maxfd           int
	setsize         int
	timeEventNextId int64
	lastTime        time.Time
	events          []*FileEvent
	fired           []*FireEvent
	apidata         interface{}

	stop bool
}

type FileProc func(eventLoop *EventLoop, fd int, clientData interface{}, mask int)

type FileEvent struct {
	mask       int
	rfileProc  FileProc
	wfileProc  FileProc
	clientData interface{}
}

type FireEvent struct {
	mask int
	fd   int
}

func CreateEventLoop(setsize int) (*EventLoop, bool) {
	el := &EventLoop{
		maxfd:           -1,
		setsize:         setsize,
		timeEventNextId: 0,
		lastTime:        time.Now(),
		events:          make([]*FileEvent, setsize),
		fired:           make([]*FireEvent, setsize),
	}
	for i := 0; i < setsize; i++ {
		fe := &FileEvent{mask: AE_NONE}
		el.events[i] = fe
		el.fired[i] = &FireEvent{}
	}
	if !apiCreate(el) {
		return nil, false
	}
	return el, true
}

func (el *EventLoop) processEvents() int {
	processed := 0
	n := apiPoll(el)
	for i := 0; i < n; i++ {
		fd := el.fired[i].fd
		fe := el.events[fd]
		mask := el.fired[i].mask
		fired := 0

		invert := fe.mask&AE_BARRIER != 0
		if !invert && fe.mask&mask&AE_READABLE != 0 {
			fe.rfileProc(el, fd, fe.clientData, mask)
			fired++
		}

		if fe.mask&mask&AE_WRITABLE != 0 {
			if fired != 0 {
				fe.wfileProc(el, fd, fe.clientData, mask)
				fired++
			}
		}

		if invert && fe.mask&mask&AE_READABLE != 0 {
			if fired != 0 {
				fe.rfileProc(el, fd, fe.clientData, mask)
				fired++
			}
		}
		processed++
	}
	return processed
}

func (el *EventLoop) CreateFileEvent(fd int, mask int, proc FileProc, clientData interface{}) error {
	if fd >= el.setsize {
		return errors.New("math result not representable")
	}

	if ok := apiAddEvent(el, fd, mask); !ok {
		return errors.New("add event failed")
	}
	fe := el.events[fd]
	fe.mask |= mask
	if mask&AE_READABLE != 0 {
		fe.rfileProc = proc
	}
	if mask&AE_WRITABLE != 0 {
		fe.wfileProc = proc
	}
	fe.clientData = clientData
	if fd > el.maxfd {
		el.maxfd = fd
	}
	return nil
}

func AeMain(el *EventLoop) {
	el.stop = false
	for !el.stop {
		el.processEvents()
	}
}
