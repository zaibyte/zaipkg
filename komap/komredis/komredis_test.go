package komredis

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zaibyte/zaipkg/uid"
	"github.com/zaibyte/zaipkg/xmath/xrand"
)

func TestSingleKMRVal(t *testing.T) {

	p := make([]byte, 44)

	oid := uid.GenRandOIDs(1)[0]
	size := xrand.Uint32()
	makeKMRVal(p, oid, size)

	kvPairs, oid2, size2 := fastParseKMRVal(p[28:])

	assert.Equal(t, uint32(1), kvPairs)
	assert.Equal(t, oid, oid2)
	assert.Equal(t, size, size2)
}

func TestInsertKeyInVal(t *testing.T) {

	p := make([]byte, 44)

	oid := uid.GenRandOIDs(1)[0]
	size := xrand.Uint32()
	makeKMRVal(p, oid, size)

	oid2, size2 := uid.GenRandOIDs(1)[0], xrand.Uint32()

	key := "1"

	unchanged, oldOID, newVal := insertKeyInVal(key, p[28:], oid2, size2)
	assert.False(t, unchanged)
	assert.Equal(t, uint64(0), oldOID)
	assert.Equal(t, 33, len(newVal))
	kvPairs, _, _ := fastParseKMRVal(newVal)
	assert.Equal(t, uint32(2), kvPairs)

	found, soid, ssize := searchKeyInVal(key, newVal, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid2, soid)
	assert.Equal(t, size2, ssize)

	unchanged, oldOID, newVal2 := insertKeyInVal(key, newVal, oid2, size2)
	assert.True(t, unchanged)
	assert.Equal(t, uint64(0), oldOID)
	assert.Equal(t, newVal, newVal2)
	assert.Equal(t, 33, len(newVal2))
	kvPairs, _, _ = fastParseKMRVal(newVal2)
	assert.Equal(t, uint32(2), kvPairs)

	unchanged, oldOID, newVal3 := insertKeyInVal(key+"2", newVal, oid2, size2)
	assert.False(t, unchanged)
	assert.Equal(t, uint64(0), oldOID)
	assert.NotEqual(t, newVal2, newVal3)
	assert.Equal(t, 33+12+4, len(newVal3))
	kvPairs, _, _ = fastParseKMRVal(newVal3)
	assert.Equal(t, uint32(3), kvPairs)

	found, soid, ssize = searchKeyInVal(key, newVal3, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid2, soid)
	assert.Equal(t, size2, ssize)
	found, soid, ssize = searchKeyInVal(key+"2", newVal3, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid2, soid)
	assert.Equal(t, size2, ssize)

	unchanged, oldOID, newVal4 := insertKeyInVal(key+"2", newVal3, oid, size)
	assert.False(t, unchanged)
	assert.Equal(t, oid2, oldOID)
	assert.NotEqual(t, newVal3, newVal4)
	assert.Equal(t, 33+12+4, len(newVal4))
	kvPairs, _, _ = fastParseKMRVal(newVal4)
	assert.Equal(t, uint32(3), kvPairs)
}

func TestDeleteKeyInVal(t *testing.T) {

	p := make([]byte, 44)

	oid := uid.GenRandOIDs(1)[0]
	size := xrand.Uint32()
	makeKMRVal(p, oid, size)

	// Add two more key.
	key1 := "1"
	oid1, size1 := uid.GenRandOIDs(1)[0], xrand.Uint32()
	_, _, valWithKey1 := insertKeyInVal(key1, p[28:], oid1, size1)
	key2 := "2"
	oid2, size2 := uid.GenRandOIDs(1)[0], xrand.Uint32()
	_, _, valWithKey2 := insertKeyInVal(key2, valWithKey1, oid2, size2)

	assert.Equal(t, 48, len(valWithKey2))

	// Found by key matched.
	kvPairs, _, _ := fastParseKMRVal(valWithKey2)
	assert.Equal(t, uint32(3), kvPairs)
	found, doid, valAfterDel := deleteKeyInVal(key1, valWithKey2, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid1, doid)
	assert.Equal(t, 48-3-12, len(valAfterDel))
	kvPairs, _, _ = fastParseKMRVal(valAfterDel)
	assert.Equal(t, uint32(2), kvPairs)

	// Found by delete first one(which has no origin key)
	found, soid, ssize := searchKeyInVal(key2+"2", valAfterDel, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid, soid)
	assert.Equal(t, size, ssize)

	found, doid, valAfterDel = deleteKeyInVal(key2+"a", valAfterDel, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid, doid)
	kvPairs, _, _ = fastParseKMRVal(valAfterDel)
	assert.Equal(t, uint32(1), kvPairs)
	assert.Equal(t, 19, len(valAfterDel))

	// Delete last key.
	found, doid, valAfterDel2 := deleteKeyInVal(key2, valAfterDel, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid2, doid)
	assert.Equal(t, 0, len(valAfterDel2))

	// Cannot find.
	kvPairs, _, _ = fastParseKMRVal(valAfterDel)
	assert.Equal(t, uint32(1), kvPairs)

	found, soid, ssize = searchKeyInVal(key2+"a", valAfterDel, kvPairs)
	assert.False(t, found)
	assert.Equal(t, uint64(0), soid)
	assert.Equal(t, uint32(0), ssize)

	found, doid, valAfterDel = deleteKeyInVal(key2+"a", valAfterDel, kvPairs)
	assert.False(t, found)
	assert.Equal(t, uint64(0), doid)
	kvPairs, _, _ = fastParseKMRVal(valAfterDel)
	assert.Equal(t, uint32(1), kvPairs)
	assert.Equal(t, 19, len(valAfterDel))

	// Add two keys back.
	_, _, valWithKey1 = insertKeyInVal(key1, p[28:], oid1, size1)
	_, _, valWithKey2 = insertKeyInVal(key2, valWithKey1, oid2, size2)

	// Delete second one.
	kvPairs, _, _ = fastParseKMRVal(valWithKey2)
	assert.Equal(t, uint32(3), kvPairs)
	found, doid, valAfterDel = deleteKeyInVal(key2, valWithKey2, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid2, doid)
	assert.Equal(t, 48-3-12, len(valAfterDel))

	// Delete first one with origin key.
	kvPairs, _, _ = fastParseKMRVal(valAfterDel)
	assert.Equal(t, uint32(2), kvPairs)
	found, doid, valAfterDel = deleteKeyInVal(key1, valAfterDel, kvPairs)
	assert.True(t, found)
	assert.Equal(t, oid1, doid)
	assert.Equal(t, 16, len(valAfterDel))
}
