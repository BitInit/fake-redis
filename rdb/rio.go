package rdb

import (
	"os"

	"github.com/BitInit/fake-redis/util"
)

type rio interface {
	read([]byte) int
	write([]byte) (int, error)
	updateCkSum([]byte) error
	getCkSum() uint64
}

type fileRio struct {
	fp       *os.File
	chckSum  uint64
	buffered int
	autosync int
}

func (r *fileRio) updateCkSum(buf []byte) error {
	r.chckSum = util.Crc64(r.chckSum, buf)
	return nil
}

func (r *fileRio) getCkSum() uint64 {
	return r.chckSum
}

func (fr *fileRio) read(buf []byte) int {
	return 0
}

func (fr *fileRio) write(buf []byte) (int, error) {
	return fr.fp.Write(buf)
}
