package main

import (
	"log"
	"strconv"
)

const OBJ_ENCODING_RAW = 0 // Raw representation
const OBJ_ENCODING_INT = 1 // Encoded as integer
const OBJ_ENCODING_HT = 2
const OBJ_ENCODING_EMBSTR = 8 // Embedded sds string encoding

const OBJ_STRING = 0 // String object.
const OBJ_LIST = 1   // List object.
const OBJ_SET = 2    // Set object.
const OBJ_ZSET = 3   // Sorted set object.
const OBJ_HASH = 4   // Hash object.

type robj struct {
	tp       uint8
	encoding uint8
	lru      uint32
	refcount int
	ptr      interface{}
}

func createStringObject(bs []byte) *robj {
	newbs := make([]byte, len(bs))
	copy(newbs, bs)
	return createObject(OBJ_STRING, newbs)
}

func createObject(tp int, ptr interface{}) *robj {
	return &robj{
		tp:       uint8(tp),
		encoding: OBJ_ENCODING_RAW,
		lru:      lruClock(),
		refcount: 1,
		ptr:      ptr,
	}
}

func sdsEncodedObject(obj *robj) bool {
	return obj.encoding == OBJ_ENCODING_RAW || obj.encoding == OBJ_ENCODING_EMBSTR
}

func getLongLongFromObject(o *robj) (int64, bool) {
	if o == nil {
		return 0, true
	}
	if sdsEncodedObject(o) {
		v := string(o.ptr.([]byte))
		if i, err := strconv.Atoi(v); err != nil {
			return 0, false
		} else {
			return int64(i), true
		}
	} else if o.encoding == OBJ_ENCODING_INT {
		v := o.ptr.(int)
		return int64(v), true
	} else {
		log.Fatalln("Unknown string encoding")
	}
	return 0, false
}

func getLongLongFromObjectOrReply(c *client, o *robj, msg string) (int64, bool) {
	val, ok := getLongLongFromObject(o)
	if !ok {
		if len(msg) > 0 {
			addReplyError(c, msg)
		} else {
			addReplyError(c, "value is not an integer or out of range")
		}
		return 0, false
	}
	return val, true
}
