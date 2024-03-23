package sds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakRoomFor(t *testing.T) {
	s := Empty()
	s.MakeRoomFor(10)
	assert.Equal(t, 20, s.alloc)
	assert.Equal(t, 0, s.len)
}

func TestCat(t *testing.T) {
	s := Empty()
	s.Cat([]byte("hello"))
	assert.Equal(t, 5, s.len)
	s.Cat([]byte(" world"))
	assert.Equal(t, 11, s.len)
}

func TestRange(t *testing.T) {
	s := Empty()
	s.Catstr("hello world")
	s.Range(2, -1)
	assert.Equal(t, 9, s.len)
	assert.Equal(t, "llo world", string(s.Buf[:s.len]))
}
