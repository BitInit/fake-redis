package main

func selectDb(c *client, id int) bool {
	if id < 0 || id >= server.dbnum {
		return false
	}
	c.db = id
	return true
}
