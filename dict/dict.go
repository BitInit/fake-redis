package dict

import (
	"math"
)

const dictHtInitialSize = 4
const dict_force_resize_ratio = 5
const dict_can_resize = true

type DictEntry struct {
	key  interface{}
	v    interface{}
	next *DictEntry
}

func (de *DictEntry) GetVal() interface{} {
	return de.v
}

type DictType struct {
	HashFunction func(key interface{}) uint64
	KeyDup       func(privdata interface{}, key interface{}) interface{}
	ValDup       func(privdata interface{}, obj interface{}) interface{}
	KeyCompare   func(privdata interface{}, key1 interface{}, key2 interface{}) int
}

type dictht struct {
	table    []*DictEntry
	size     uint64
	sizemask uint64
	used     uint64
}

type Dict struct {
	dictType  *DictType
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

func Create(tp *DictType, privdata interface{}) *Dict {
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
	return d.dictType.HashFunction(key)
}

func (d *Dict) compareKeys(key1 interface{}, key2 interface{}) int {
	return d.dictType.KeyCompare(d.privdata, key1, key2)
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
	n.table = make([]*DictEntry, realsize)
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

	if d.ht[0].used >= d.ht[0].size && (dict_can_resize || d.ht[0].used/d.ht[0].size > dict_force_resize_ratio) {
		return d.expand(d.ht[0].used * 2)
	}
	return true
}

func (d *Dict) keyIndex(key interface{}, hash uint64, existing **DictEntry) int64 {
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
			if d.compareKeys(key, he.key) == 0 {
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

func (d *Dict) rehash(n int) {
	empty_visits := n * 10
	if !d.isRehashing() {
		return
	}

	for ; n > 0 && d.ht[0].used != 0; n-- {
		for d.ht[0].table[d.rehashidx] == nil {
			d.rehashidx++
			empty_visits--
			if empty_visits == 0 {
				return
			}
		}

		de := d.ht[0].table[d.rehashidx]
		for de != nil {
			nextde := de.next

			h := d.hashKey(de.key) & d.ht[1].sizemask
			de.next = d.ht[1].table[h]
			d.ht[1].table[h] = de
			d.ht[0].used--
			d.ht[1].used++
			de = nextde
		}
		d.ht[0].table[d.rehashidx] = nil
		d.rehashidx++
	}

	if d.ht[0].used == 0 {
		d.ht[0].table = nil
		d.ht[0] = d.ht[1]
		initDictht(&d.ht[1])
		d.rehashidx = -1
	}
}

func (d *Dict) rehashStep() {
	if d.iterators != 0 {
		return
	}
	d.rehash(1)
}

func (d *Dict) addRaw(key interface{}, existing **DictEntry) *DictEntry {
	if d.isRehashing() {
		d.rehashStep()
	}

	index := d.keyIndex(key, d.hashKey(key), existing)
	if index == -1 {
		return nil
	}

	ht := &d.ht[0]
	if d.isRehashing() {
		ht = &d.ht[1]
	}
	entry := &DictEntry{
		next: ht.table[index],
	}
	ht.used++
	ht.table[index] = entry
	d.setKey(entry, key)
	return entry
}

func (d *Dict) setKey(de *DictEntry, key interface{}) {
	if d.dictType.KeyDup != nil {
		de.key = d.dictType.KeyDup(d.privdata, key)
		return
	}
	de.key = key
}

func (d *Dict) SetEntryVal(de *DictEntry, val interface{}) {
	if d.dictType.ValDup != nil {
		de.v = d.dictType.ValDup(d.privdata, val)
		return
	}
	de.v = val
}

func (d *Dict) Add(key interface{}, val interface{}) bool {
	entry := d.addRaw(key, nil)

	if entry == nil {
		return false
	}
	d.SetEntryVal(entry, val)
	return true
}

func (d *Dict) Find(key interface{}) *DictEntry {
	if d.ht[0].used+d.ht[1].used == 0 {
		return nil
	}

	if d.isRehashing() {
		d.rehashStep()
	}
	h := d.hashKey(key)
	for table := 0; table <= 1; table++ {
		idx := h & d.ht[table].sizemask
		he := d.ht[table].table[idx]
		for he != nil {
			if d.compareKeys(key, he.key) == 0 {
				return he
			}
			he = he.next
		}
		if !d.isRehashing() {
			return nil
		}
	}
	return nil
}

func (d *Dict) GetVal(key interface{}) interface{} {
	de := d.Find(key)
	if de == nil {
		return nil
	}
	return de.v
}

func (d *Dict) GetSignedIntVal(key interface{}) (int64, bool) {
	de := d.Find(key)
	if de == nil {
		return 0, false
	}
	if v, ok := de.v.(int64); !ok {
		return 0, false
	} else {
		return v, true
	}
}

func (d *Dict) SetSignedIntVal(key interface{}, val int64) {
	var existing *DictEntry
	de := d.addRaw(key, &existing)
	if de == nil {
		existing.v = val
	} else {
		de.v = val
	}
}

func (d *Dict) genericDelete(key interface{}) *DictEntry {
	if d.ht[0].used == 0 && d.ht[1].used == 0 {
		return nil
	}

	if d.isRehashing() {
		d.rehashStep()
	}
	h := d.hashKey(key)
	for talbe := 0; talbe <= 1; talbe++ {
		idx := h & d.ht[talbe].sizemask
		he := d.ht[talbe].table[idx]
		var prevHe *DictEntry = nil
		for he != nil {
			if d.compareKeys(he.key, key) == 0 {
				if prevHe != nil {
					prevHe.next = he.next
				} else {
					d.ht[talbe].table[idx] = he.next
				}
				d.ht[talbe].used--
				return he
			}
			prevHe = he
			he = he.next
		}

		if !d.isRehashing() {
			break
		}
	}
	return nil
}

func (d *Dict) Delete(key interface{}) bool {
	return d.genericDelete(key) != nil
}
