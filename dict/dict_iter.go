package dict

type DictIterator struct {
	d                *Dict
	index            int
	table            int
	safe             bool
	entry, nextEntry *DictEntry
	// fingerprint      int64
}

func (di *DictIterator) Next() *DictEntry {
	for {
		if di.entry == nil {
			ht := &di.d.ht[di.table]
			if di.index == -1 && di.table == 0 {
				if di.safe {
					di.d.iterators++
				}
				// else {
				// di.fingerprint =
				// }
			}
			di.index++
			if di.index >= int(ht.size) {
				if di.d.isRehashing() && di.table == 0 {
					di.table++
					di.index = 0
					ht = &di.d.ht[1]
				} else {
					break
				}
			}
			di.entry = ht.table[di.index]
		} else {
			di.entry = di.nextEntry
		}
		if di.entry != nil {
			di.nextEntry = di.entry.next
			return di.entry
		}
	}
	return nil
}
