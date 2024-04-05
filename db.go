package main

import (
	"time"

	"github.com/BitInit/fake-redis/dict"
)

type redisDb struct {
	id      int        // Database ID
	dict    *dict.Dict // The keyspace for this DB
	expires *dict.Dict
}

func selectDb(c *client, id int) bool {
	if id < 0 || id >= server.dbnum {
		return false
	}
	c.db = server.db[id]
	return true
}

func getExpire(db *redisDb, key *robj) int64 {
	v, ok := db.expires.GetSignedIntVal(key.ptr)
	if !ok {
		return -1
	}
	return v
}

func keyIsExpired(db *redisDb, key *robj) bool {
	when := getExpire(db, key)
	if when < 0 {
		return false
	}

	now := time.Now()
	return now.UnixMilli() > when
}

func lookupKey(db *redisDb, key *robj, flags int) *robj {
	val := db.dict.GetVal(key.ptr)
	if val == nil {
		return nil
	}
	// set lru clock
	return val.(*robj)
}

func lookupKeyWrite(db *redisDb, key *robj) *robj {
	expireIfNeeded(db, key)
	return lookupKey(db, key, 0)
}

// The return value of the function is true if the key is still valid,
// otherwise the function returns false if the key is expired.
func expireIfNeeded(db *redisDb, key *robj) bool {
	if !keyIsExpired(db, key) {
		return true
	}

	// delete the key
	db.expires.Delete(key.ptr)
	db.dict.Delete(key.ptr)
	return false
}

func dbAdd(db *redisDb, key, val *robj) {

	db.dict.Add(key.ptr, val)
}

func dbOverwirte(db *redisDb, key, val *robj) {
	old := db.dict.Find(key.ptr)
	db.dict.SetEntryVal(old, val)
}

func setKey(db *redisDb, key, val *robj) {
	if lookupKeyWrite(db, key) == nil {
		dbAdd(db, key, val)
	} else {
		dbOverwirte(db, key, val)
	}
}

func setExpire(c *client, db *redisDb, key *robj, when int64) {
	db.expires.SetSignedIntVal(key.ptr, when)
}

func lookupKeyReadWithFlags(db *redisDb, key *robj, flags int) *robj {
	if !expireIfNeeded(db, key) {
		return nil
	}
	val := lookupKey(db, key, flags)
	return val
}

func lookupKeyReadOrReply(c *client, key, reply *robj) *robj {
	o := lookupKeyReadWithFlags(c.db, key, 0)
	if o == nil {
		addReply(c, reply)
	}
	return o
}
