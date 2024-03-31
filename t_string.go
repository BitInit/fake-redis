package main

func getCommand(c *client) {
	addReplyString(c, "+OK for get\r\n")
}

func setCommand(c *client) {
	addReplyString(c, "+OK for set\r\n")
}

func setnxCommand(c *client) {
	addReplyString(c, "+OK for setnx\r\n")
}

func delCommand(c *client) {
	addReplyString(c, "+OK for del\r\n")
}
