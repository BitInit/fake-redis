package main

const OBJ_ENCODING_RAW = 0 // Raw representation
const OBJ_ENCODING_INT = 1 // Encoded as integer
const OBJ_ENCODING_HT = 2

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

func createObject(tp int, ptr interface{}) *robj {
	return &robj{
		tp:       uint8(tp),
		encoding: OBJ_ENCODING_RAW,
		lru:      lruClock(),
		refcount: 1,
		ptr:      ptr,
	}
}
