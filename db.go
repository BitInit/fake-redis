package main

import "github.com/BitInit/fake-redis/dict"

type redisDb struct {
	id      int        // Database ID
	dict    *dict.Dict // The keyspace for this DB
	expires *dict.Dict
}

func selectDb(c *client, id int) bool {
	if id < 0 || id >= server.dbnum {
		return false
	}
	c.db = id
	return true
}
