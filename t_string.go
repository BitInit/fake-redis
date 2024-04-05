package main

import "time"

const OBJ_SET_NO_FLAGS = 0
const OBJ_SET_NX = (1 << 0)
const OBJ_SET_XX = (1 << 1)
const OBJ_SET_EX = (1 << 2)
const OBJ_SET_PX = (1 << 3)

func setGenericCommand(c *client, flags int, key, val, expire *robj, unit int) {
	var milliseconds int64 = 0
	if expire != nil {
		var ok bool
		milliseconds, ok = getLongLongFromObjectOrReply(c, expire, "")
		if !ok {
			return
		}
		if milliseconds < 0 {
			addReplyErrorFormat(c, "invalid expire time in %s", c.cmd.name)
			return
		}
		if unit == UNIT_SECONDS {
			milliseconds *= 1000
		}
	}

	if ((flags&OBJ_SET_NX) != 0 && lookupKeyWrite(c.db, key) != nil) ||
		(flags&OBJ_SET_XX != 0 && lookupKeyWrite(c.db, key) == nil) {
		addReply(c, shared.nullbulk)
		return
	}
	setKey(c.db, key, val)
	server.dirty++
	if expire != nil {
		setExpire(c, c.db, key, time.Now().UnixMilli()+milliseconds)
	}
	addReply(c, shared.ok)
}

func getGenericCommand(c *client) {
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)
	if o == nil {
		return
	}
	addReplyBulk(c, o)
}

func getCommand(c *client) {
	getGenericCommand(c)
}

// SET key value [NX] [XX] [EX <seconds>] [PX <milliseconds>]
func setCommand(c *client) {
	var expire *robj = nil
	var flags int = OBJ_SET_NO_FLAGS
	var unit int = UNIT_SECONDS

	for j := 3; j < c.argc; j++ {
		a := c.argv[j].ptr.([]byte)
		var next *robj = nil
		if j < c.argc-1 {
			next = c.argv[j+1]
		}

		if (a[0] == 'n' || a[0] == 'N') &&
			(a[1] == 'x' || a[1] == 'X') &&
			len(a) == 2 && (flags&OBJ_SET_XX) == 0 {
			flags |= OBJ_SET_NX
		} else if (a[0] == 'x' || a[0] == 'X') &&
			(a[1] == 'x') || a[1] == 'X' &&
			len(a) == 2 && (flags&OBJ_SET_NX) == 0 {
			flags |= OBJ_SET_XX
		} else if (a[0] == 'e' || a[0] == 'E') &&
			(a[1] == 'x' || a[1] == 'X') &&
			len(a) == 2 && (flags&OBJ_SET_PX) == 0 &&
			next != nil {
			flags |= OBJ_SET_EX
			unit = UNIT_SECONDS
			expire = next
			j++
		} else if (a[0] == 'p' || a[0] == 'P') &&
			(a[1] == 'x' || a[1] == 'X') &&
			len(a) == 2 && (flags&OBJ_SET_EX) == 0 &&
			next != nil {
			flags = OBJ_SET_PX
			unit = UNIT_MILLISECONDS
			expire = next
			j++
		} else {
			addReply(c, shared.syntaxerr)
			return
		}
	}

	setGenericCommand(c, flags, c.argv[1], c.argv[2], expire, unit)
}

func setnxCommand(c *client) {
	addReplyString(c, "+OK for setnx\r\n")
}

func delCommand(c *client) {
	addReplyString(c, "+OK for del\r\n")
}
