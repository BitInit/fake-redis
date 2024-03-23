package sds

const sdsMaxPrealloc = 1024 * 1024

type Sds struct {
	alloc int
	len   int
	Buf   []byte
}

func Empty() *Sds {
	return &Sds{
		alloc: 0,
		len:   0,
	}
}

func (s *Sds) Cat(t []byte) *Sds {
	addlen := len(t)
	s.MakeRoomFor(addlen)
	copy(s.Buf[s.len:], t)
	s.len += addlen
	return s
}

func (s *Sds) Catstr(str string) *Sds {
	return s.Cat([]byte(str))
}

func (s *Sds) Len() int {
	return s.len
}

func (s *Sds) IncLen(len int) {
	s.len += len
}

func (s *Sds) Isempty() bool {
	return s.len == 0
}

func (s *Sds) MakeRoomFor(addlen int) *Sds {
	avail := s.alloc - s.len
	if avail >= addlen {
		return s
	}
	newlen := s.len + addlen
	if newlen < sdsMaxPrealloc {
		newlen *= 2
	} else {
		newlen += sdsMaxPrealloc
	}
	newBuf := make([]byte, newlen)
	copy(newBuf, s.Buf[:s.len])
	s.Buf = newBuf
	s.alloc = cap(s.Buf)
	return s
}

func (s *Sds) Range(start, end int) *Sds {
	oldlen := s.len
	if oldlen == 0 || start >= oldlen {
		return s
	}
	if end > oldlen || end < 0 {
		end = oldlen
	}
	copy(s.Buf, s.Buf[start:end])
	s.len = end - start
	return s
}
