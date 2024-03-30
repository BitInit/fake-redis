package ae

import (
	"errors"
	"time"
)

const AE_NONE = 0
const AE_READABLE = 1
const AE_WRITABLE = 2
const AE_BARRIER = 4

type beforesleepProc func()

type EventLoop struct {
	maxfd           int
	setsize         int
	timeEventNextId int64
	lastTime        time.Time
	events          []*FileEvent
	fired           []*FireEvent
	apidata         interface{}
	stop            bool

	beforesleep beforesleepProc
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

func CreateEventLoop(setsize int) (*EventLoop, error) {
	el := &EventLoop{
		maxfd:           -1,
		setsize:         setsize,
		timeEventNextId: 0,
		stop:            false,
		lastTime:        time.Now(),
		events:          make([]*FileEvent, setsize),
		fired:           make([]*FireEvent, setsize),
	}
	for i := 0; i < setsize; i++ {
		fe := &FileEvent{mask: AE_NONE}
		el.events[i] = fe
		el.fired[i] = &FireEvent{}
	}
	if err := apiCreate(el); err != nil {
		return nil, err
	}
	return el, nil
}

func (el *EventLoop) SetBeforeSleepProc(proc beforesleepProc) {
	el.beforesleep = proc
}

func (el *EventLoop) processEvents() int {
	processed := 0
	n := apiPoll(el, nil)
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

	if err := apiAddEvent(el, fd, mask); err != nil {
		return err
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

func (el *EventLoop) DeleteFileEvent(fd int, mask int) {
	if fd >= el.setsize {
		return
	}

	fe := el.events[fd]
	if fe.mask == AE_NONE {
		return
	}

	apiDelEvent(el, fd, mask)
	fe.mask = AE_NONE
	if fd == el.maxfd && fe.mask == AE_NONE {
		j := 0
		for j = el.maxfd - 1; j >= 0; j-- {
			if el.events[j].mask != AE_NONE {
				break
			}
		}
		el.maxfd = j
	}
}

func AeMain(el *EventLoop) {
	el.stop = false
	for !el.stop {
		if el.beforesleep != nil {
			el.beforesleep()
		}
		el.processEvents()
	}
}
