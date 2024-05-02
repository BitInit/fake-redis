package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/BitInit/fake-redis/rdb"
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

	return nil
}

func rdbSaveInfoAuxFields(r *rdb.Rdb) error {
	redisBits := unsafe.Sizeof(r)

	if err := r.SaveAuxField("redis-ver", VERSION); err != nil {
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
	return r.SaveAuxField(key, v)
}
