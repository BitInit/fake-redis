package dict

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDictNextPower(t *testing.T) {
	assert.Equal(t, uint64(4), dictNextPower(4))
	assert.Equal(t, uint64(8), dictNextPower(5))
	assert.Equal(t, uint64(1024), dictNextPower(555))
}

var uint64DictType *DictType = &DictType{
	HashFunction: func(key interface{}) uint64 {
		return key.(uint64)
	},
	KeyCompare: func(privdata, key1, key2 interface{}) int {
		i1 := key1.(uint64)
		i2 := key2.(uint64)
		return int(i1) - int(i2)
	},
}

func TestAdd(t *testing.T) {
	d := Create(uint64DictType, nil)
	d.Add(uint64(1), "v1")
	assert.Equal(t, uint64(1), d.ht[0].used)

	d.Add(uint64(5), "v5")
	assert.Equal(t, uint64(2), d.ht[0].used)
}
