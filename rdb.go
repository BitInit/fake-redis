package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/BitInit/fake-redis/dict"
	"github.com/BitInit/fake-redis/rdb"
)

const (
	rdbOpCodeAux          = 250
	rdbOpCodeResizeDB     = 251
	rdbOpCodeExpireTimeMS = 252
	rdbOpCodeSelectDB     = 254
	rdbOpCodeEOF          = 255
)

const (
	rdbTypeString = 0
)

const RDB_VERSION = 9
const RDB_OPCODE_AUX = 250

const RDB_6BITLEN = 0
const RDB_14BITLEN = 1
const RDB_32BITLEN = 0x80
const RDB_64BITLEN = 0x81

func saveCommand(c *client) {
	if server.rdb_child_pid != -1 {
		addReplyError(c, "Background save already in progress")
	}

	if rdbSave(server.rdb_filename) {
		addReply(c, shared.ok)
	} else {
		addReply(c, shared.err)
	}
}

func rdbSave(filename string) bool {
	tmpfile := fmt.Sprintf("temp-%d.rdb", syscall.Getpid())
	f, err := os.Create(tmpfile)
	if err != nil {
		log.Printf("Failed opening the RDB file %s\n", tmpfile)
		return false
	}
	defer f.Close()

	r := rdb.InitWithFile(f)
	if err := rdbSaveRio(r); err != nil {
		return false
	}

	if err := os.Rename(tmpfile, filename); err != nil {
		// TODO unlink
		return false
	}

	return true
}

func rdbSaveRio(r *rdb.Rdb) error {
	magic := fmt.Sprintf("REDIS%04d", RDB_VERSION)
	if _, err := r.WriteRaw([]byte(magic)); err != nil {
		return err
	}
	if err := rdbSaveInfoAuxFields(r); err != nil {
		return err
	}

	for i := 0; i < server.dbnum; i++ {
		db := server.db[i]
		saveDb(r, i, db.dict, db.expires)
	}
	return nil
}

func saveDb(r *rdb.Rdb, i int, kv *dict.Dict, expires *dict.Dict) error {
	dbSize := kv.Size()
	expiresSize := expires.Size()
	if dbSize == 0 {
		return nil
	}
	if err := r.SaveType(rdbOpCodeSelectDB); err != nil {
		return err
	}
	if _, err := r.SaveLen(uint64(i)); err != nil {
		return err
	}

	if err := r.SaveType(rdbOpCodeResizeDB); err != nil {
		return err
	}
	if _, err := r.SaveLen(uint64(dbSize)); err != nil {
		return err
	}
	if _, err := r.SaveLen(uint64(expiresSize)); err != nil {
		return err
	}

	di := kv.GetSafeInterator()
	for de := di.Next(); de != nil; de = di.Next() {
		key := de.GetKey().([]byte)
		val := de.GetVal().(*robj)
		expire := expires.GetVal(key)

		if expire != nil {
			if err := r.SaveType(rdbOpCodeExpireTimeMS); err != nil {
				return err
			}
			et := expire.(int64)
			if _, err := r.SaveLen(uint64(et)); err != nil {
				return err
			}
		}
		if err := saveObjectType(r, val); err != nil {
			return err
		}
		if err := r.SaveRawString(string(key)); err != nil {
			return err
		}
		if err := saveObject(r, val); err != nil {
			return err
		}

	}

	if err := r.SaveType(rdbOpCodeEOF); err != nil {
		return err
	}
	if err := r.SaveCheckSum(); err != nil {
		return err
	}
	return nil
}

func rdbSaveInfoAuxFields(r *rdb.Rdb) error {
	redisBits := unsafe.Sizeof(r)

	if err := saveAuxField(r, "redis-ver", VERSION); err != nil {
		return err
	}
	if err := rdbSaveAuxFieldStrInt(r, "redis-bits", int(redisBits)); err != nil {
		return err
	}
	if err := rdbSaveAuxFieldStrInt(r, "ctime", int(time.Now().UnixMilli())); err != nil {
		return err
	}
	return nil
}

func rdbSaveAuxFieldStrInt(r *rdb.Rdb, key string, val int) error {
	v := strconv.Itoa(val)
	return saveAuxField(r, key, v)
}

func saveAuxField(r *rdb.Rdb, key string, val string) error {
	if err := r.SaveType(rdbOpCodeAux); err != nil {
		return err
	}
	if err := r.SaveRawString(key); err != nil {
		return err
	}
	if err := r.SaveRawString(val); err != nil {
		return err
	}
	return nil
}

func saveObjectType(r *rdb.Rdb, o *robj) error {
	switch o.tp {
	case OBJ_STRING:
		return r.SaveType(rdbTypeString)
		// other type
	}
	return nil
}

func saveObject(r *rdb.Rdb, val *robj) error {
	if val.tp == OBJ_STRING {
		return r.SaveRawString(string(val.ptr.([]byte)))
	}
	return fmt.Errorf("no support type")
}
