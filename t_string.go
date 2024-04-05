package main

import "strconv"

func getCommand(c *client) {
	d := c.db.dict
	v := d.GetVal(c.argv[1].ptr)
	if v == nil {
		addReply(c, shared.nullbulk)
		return
	}
	o := v.(*robj)
	val := o.ptr.([]byte)
	addReplyString(c, "$"+strconv.Itoa(len(val))+"\r\n")
	addReplyString(c, string(val)+"\r\n")
}

func setCommand(c *client) {
	d := c.db.dict
	v := d.GetVal(c.argv[1].ptr)
	if v == nil {
		d.Add(c.argv[1].ptr, c.argv[2])
		addReplyString(c, "+OK\r\n")
		return
	}
	o := v.(*robj)
	o.ptr = c.argv[2].ptr
	addReplyString(c, "+OK\r\n")
}

func setnxCommand(c *client) {
	addReplyString(c, "+OK for setnx\r\n")
}

func delCommand(c *client) {
	addReplyString(c, "+OK for del\r\n")
}
