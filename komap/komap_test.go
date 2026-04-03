package komap

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/zaibyte/zaipkg/komap/komredis"

	"github.com/stretchr/testify/assert"
	"github.com/zaibyte/zaipkg/uid"
)

func TestKomapRedis(t *testing.T) {

	k, err := komredis.NewKomapRedis("redis://127.0.0.1:6379/11")
	if err != nil {
		t.Skipf("redis is not available: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err = k.RDB.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %s", err)
	}
	k.RDB.FlushDB(ctx)
	runTestKOMap(t, k)
}

// TODO need more testing.
func runTestKOMap(t *testing.T, k KoMap) {

	key := "1"
	oid := uid.GenRandOIDs(1)[0]

	size := uid.GetGrains(oid)*uid.GrainSize - uint32(rand.Int31n(4096))

	oldOID, err := k.Insert(context.Background(), key, oid, size)
	assert.Nil(t, err)
	assert.Equal(t, oldOID, uint64(0))

	oldOID, err = k.Insert(context.Background(), key, oid, size)
	assert.Nil(t, err)
	assert.Equal(t, oldOID, uint64(0))

	oid2, size2, err := k.Search(context.Background(), key)
	assert.Nil(t, err)
	assert.Equal(t, oid, oid2)
	assert.Equal(t, size, size2)

	oid2, err = k.Delete(context.Background(), key)
	assert.Nil(t, err)
	assert.Equal(t, oid, oid2)
}
