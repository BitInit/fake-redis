package sds

import (
	"testing"
)

func TestSdsAlloc(t *testing.T) {
	var s Sds
	s = append(s, []byte("abc")...)
	println(len(s), cap(s))
}
