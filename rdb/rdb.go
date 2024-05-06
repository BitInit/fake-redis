package rdb

import (
	"encoding/binary"
	"math"
	"os"
	"strconv"
)

const (
	rdbEncVal   = 3
	rdb6BitLen  = 0
	rdb14BitLen = 1
	rdb32BitLen = 0x80
	rdb64BitLen = 0x81

	rdbEncInt8  = 0
	rdbEncInt16 = 1
	rdbEncInt32 = 2
	rdbEncLzf   = 3
)

func encodeInteger(v int) []byte {
	if v >= -(1<<7) && v <= (1<<7)-1 {
		return []byte{
			(rdbEncVal << 6) | rdbEncInt8,
			byte(v & 0xFF),
		}
	} else if v >= -(1<<15) && v <= (1<<15)-1 {
		return []byte{
			(rdbEncVal << 6) | rdbEncInt16,
			byte(v & 0xFF),
			byte((v >> 8) & 0xFF),
		}
	} else if v >= -(1<<31) && v <= (1<<31)-1 {
		return []byte{
			(rdbEncVal << 6) | rdbEncInt32,
			byte(v & 0xFF),
			byte(v>>8) & 0xFF,
			byte(v>>16) & 0xFF,
			byte(v>>24) & 0xFF,
		}
	}
	return nil
}

func tryIntegerEncoding(s string) []byte {
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return encodeInteger(v)
}

type Rdb struct {
	_rio rio
}

func InitWithFile(fp *os.File) *Rdb {
	return &Rdb{
		_rio: &fileRio{
			fp:       fp,
			chckSum:  0,
			buffered: 0,
			autosync: 0,
		},
	}
}

func (r *Rdb) WriteRaw(buf []byte) (int, error) {
	if err := r._rio.updateCkSum(buf); err != nil {
		return 0, err
	}
	n, err := r._rio.write(buf)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (r *Rdb) SaveType(tp byte) error {
	if _, err := r.WriteRaw([]byte{tp}); err != nil {
		return err
	}
	return nil
}

func (r *Rdb) SaveRawString(s string) error {
	l := len(s)
	if l <= 11 {
		buf := tryIntegerEncoding(s)
		if len(buf) > 0 {
			if _, err := r.WriteRaw(buf); err == nil {
				return nil
			}
		}
	}

	if _, err := r.SaveLen(uint64(l)); err != nil {
		return err
	}
	if l > 0 {
		if _, err := r.WriteRaw([]byte(s)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Rdb) SaveLen(l uint64) (int, error) {
	var buf []byte
	if l < (1 << 6) {
		buf = []byte{
			byte(l&0xFF | (rdb6BitLen << 6)),
		}
	} else if l < (1 << 14) {
		buf = []byte{
			byte((l>>8)&0xFF | (rdb14BitLen << 6)),
			byte(l & 0xFF),
		}
	} else if l <= math.MaxUint32 {
		buf = []byte{rdb32BitLen}
		buf = binary.BigEndian.AppendUint64(buf, uint64(l))
	} else {
		buf = []byte{rdb64BitLen}
		buf = binary.BigEndian.AppendUint64(buf, uint64(l))
	}
	return r.WriteRaw(buf)
}

func (r *Rdb) SaveCheckSum() error {
	chSum := binary.LittleEndian.AppendUint64([]byte{}, r._rio.getCkSum())
	if _, err := r.WriteRaw(chSum); err != nil {
		return err
	}
	return nil
}
