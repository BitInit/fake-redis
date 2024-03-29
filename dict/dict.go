package dict

import (
	"math"
)

const dictHtInitialSize = 4

type dictEntry struct {
	key  interface{}
	v    interface{}
	next *dictEntry
}

type DictType interface {
	hashFunction(key interface{}) uint64
	keyDup(privdata interface{}, key interface{})
	valDup(privdata interface{}, obj interface{})
	keyCompare(privdata interface{}, key1 interface{}, key2 interface{}) int
}

type dictht struct {
	table    []*dictEntry
	size     uint64
	sizemask uint64
	used     uint64
}

type Dict struct {
	dictType  DictType
	privdata  interface{}
	ht        [2]dictht
	rehashidx int64
	iterators uint64
}

func initDictht(ht *dictht) {
	ht.table = nil
	ht.size = 0
	ht.sizemask = 0
	ht.used = 0
}

func Create(tp DictType, privdata interface{}) *Dict {
	dict := &Dict{
		dictType:  tp,
		privdata:  privdata,
		rehashidx: -1,
		iterators: 0,
	}
	initDictht(&dict.ht[0])
	initDictht(&dict.ht[1])
	return dict
}

func (d *Dict) isRehashing() bool {
	return d.rehashidx != -1
}

func (d *Dict) hashKey(key interface{}) uint64 {
	return d.dictType.hashFunction(key)
}

func (d *Dict) compareKeys(key1 interface{}, key2 interface{}) int {
	return d.dictType.keyCompare(d.privdata, key1, key2)
}

func dictNextPower(size uint64) uint64 {
	var i uint64 = dictHtInitialSize
	if size > math.MaxInt64 {
		return math.MaxInt64 + 1
	}
	for {
		if i >= size {
			return i
		}
		i *= 2
	}
}

func (d *Dict) expand(size uint64) bool {
	if d.isRehashing() || d.ht[0].used > size {
		return false
	}
	realsize := dictNextPower(size)

	if realsize == d.ht[0].size {
		return false
	}

	var n dictht
	n.size = realsize
	n.sizemask = realsize - 1
	n.table = make([]*dictEntry, realsize)
	n.used = 0

	if d.ht[0].table == nil {
		d.ht[0] = n
		return true
	}
	d.ht[1] = n
	d.rehashidx = 0
	return true
}

func (d *Dict) expandIfNeeded() bool {
	if d.isRehashing() {
		return true
	}
	if d.ht[0].size == 0 {
		return d.expand(dictHtInitialSize)
	}

	return false
}

func (d *Dict) keyIndex(key interface{}, hash uint64, existing **dictEntry) int64 {
	if existing != nil {
		*existing = nil
	}

	if !d.expandIfNeeded() {
		return -1
	}
	var idx uint64
	for tb := 0; tb <= 1; tb++ {
		idx = hash & d.ht[tb].sizemask
		he := d.ht[tb].table[idx]
		for he != nil {
			if key == he.key || d.compareKeys(key, he.key) == 0 {
				if existing != nil {
					*existing = he
				}
				return -1
			}
			he = he.next
		}
		if !d.isRehashing() {
			break
		}
	}
	return int64(idx)
}

func (d *Dict) addRaw(key interface{}, existing **dictEntry) *dictEntry {
	if d.isRehashing() {
		// do rehashing step
	}
	return nil
}

func (d *Dict) Add(key interface{}, val interface{}) {

}
