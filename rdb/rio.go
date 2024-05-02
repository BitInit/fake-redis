package rdb

import "os"

type rio interface {
	read([]byte) int
	write([]byte) (int, error)
	updateCkSum([]byte)
}

type fileRio struct {
	fp       *os.File
	buffered int
	autosync int
}

func (r *fileRio) updateCkSum(buf []byte) {}

func (fr *fileRio) read(buf []byte) int {
	return 0
}

func (fr *fileRio) write(buf []byte) (int, error) {
	return fr.fp.Write(buf)
}
