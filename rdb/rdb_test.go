package rdb

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type memRio struct {
	buf []byte
	pos int
}

func (mr *memRio) updateCkSum(buf []byte) error { return nil }

func (mr *memRio) getCkSum() uint64 {
	return 0
}

func (mr *memRio) read(buf []byte) int {
	l := len(buf)
	if l <= 0 {
		return 0
	}
	cl := copy(buf, mr.buf[mr.pos:])
	mr.pos += cl
	return cl
}

func (mr *memRio) write(buf []byte) (int, error) {
	mr.buf = append(mr.buf, buf...)
	return len(buf), nil
}

func (mr *memRio) clean() {
	mr.buf = make([]byte, 0)
	mr.pos = 0
}

func TestSaveLen(t *testing.T) {
	mr := &memRio{pos: 0}
	r := &Rdb{_rio: mr}
	r.SaveLen(31)
	assert.Equal(t, 1, len(mr.buf))
	assert.Equal(t, int(31), int(mr.buf[0]))

	mr.clean()
	r.SaveLen(64)
	assert.Equal(t, 2, len(mr.buf))
	assert.Equal(t, int(1<<6), int(mr.buf[0]))

	mr.clean()
	r.SaveLen(math.MaxUint32 + 1)
	assert.Equal(t, 9, len(mr.buf))
	assert.Equal(t, int(0x01), int(mr.buf[4]))
}
