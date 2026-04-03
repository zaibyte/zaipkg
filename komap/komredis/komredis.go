package komredis

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/zaibyte/zaipkg/xbytes"

	"github.com/zaibyte/zaipkg/orpc"
	"github.com/zaibyte/zaipkg/xredis"
	"github.com/zaibyte/zaipkg/xstrconv"

	"github.com/go-redis/redis/v8"
	"github.com/zeebo/xxh3"
)

// KomapRedis implements KoMap using redis as backend database.
// See https://github.com/zaibyte/zaifs/issues/58 for the design.
//
// struct in redis:
// key: xxh3(key_origin) % bktCnt
// field: hash1(key_origin)
// val:
// If only one pair:
// | 1(4bytes) | oid(8bytes) | size(4bytes) |
// If more than one pairs:
// | 2(4bytes) | oid0(8bytes) | size0(4bytes) | ... | 0(2bytes) | key1_origin_len(2bytes) | key1_origin | ... |
// The first key_origin_len must be 0.
//
// We only keep origin_key when there are more than one key-oid pairs have the same key & field.
//
// key: redis key
// origin_key: the real key for juiceFS.
type KomapRedis struct {
	txlocks [1024]sync.Mutex // Pessimistic locks to reduce conflict on Redis
	RDB     *redis.Client
}

// NewKomapRedis creates a KomapRedis.
func NewKomapRedis(url string) (*KomapRedis, error) {
	rdb, err := xredis.NewClient(url)
	if err != nil {
		return nil, err
	}

	r := &KomapRedis{
		RDB: rdb,
	}
	return r, nil
}

// 1 Million. If total key:object pairs is < 500 millions, we'll use ziplist for saving half of memory.
// See https://github.com/zaibyte/zaifs/issues/58#issuecomment-978 for details.
const bktCnt = 1 << 20

// Insert inserts key:<oid>_<size> pair into redis.
func (r *KomapRedis) Insert(ctx context.Context, key string, oid uint64, size uint32) (oldOID uint64, err error) {

	buf := xbytes.GetBytes(44) // 28(key & field) + 16(cnt, oid, size) val.
	defer xbytes.PutBytes(buf)

	rKey, field := getKeyHashes(key)

	binary.LittleEndian.PutUint64(buf, rKey)
	copy(buf[8:], field[:20])

	err = r.txn(ctx, func(tx *redis.Tx) error {

		val, err2 := tx.HGet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28])).Bytes()
		if err2 != nil {
			if err2 == redis.Nil {
				makeKMRVal(buf, oid, size)
				_, err2 = tx.HSet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28]), buf[28:]).Result()
			}
			return err2
		}

		// Already has same key & field, adding new and updating the val.
		// It may cause "two" pair actually have the same key_origin, but it's okay,
		// because the low collision.
		unchanged, oid2, newVal := insertKeyInVal(key, val, oid, size)
		if unchanged {
			return nil
		}
		oldOID = oid2
		_, err2 = tx.HSet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28]), newVal).Result()

		return err2
	}, key)

	return oldOID, err
}

func (r *KomapRedis) Search(ctx context.Context, key string) (oid uint64, size uint32, err error) {

	buf := xbytes.GetBytes(28) // 28(key & field)
	defer xbytes.PutBytes(buf)

	rKey, field := getKeyHashes(key)

	binary.LittleEndian.PutUint64(buf, rKey)
	copy(buf[8:], field[:20])

	val, err := r.RDB.HGet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28])).Bytes()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, orpc.ErrNotFound
		}
		return 0, 0, err
	}

	kvPairs, oidFirst, sizeFirst := fastParseKMRVal(val)
	if kvPairs == 1 {
		return oidFirst, sizeFirst, nil
	}

	// If there are more than one key_origins share the same key & field.
	found, oid, size := searchKeyInVal(key, val, kvPairs)
	if !found {
		return 0, 0, orpc.ErrNotFound
	}

	return oid, size, nil
}

func (r *KomapRedis) Delete(ctx context.Context, key string) (oid uint64, err error) {

	buf := xbytes.GetBytes(28) // 28(key & field)
	defer xbytes.PutBytes(buf)

	rKey, field := getKeyHashes(key)

	binary.LittleEndian.PutUint64(buf, rKey)
	copy(buf[8:], field[:20])

	err = r.txn(ctx, func(tx *redis.Tx) error {
		val, err2 := tx.HGet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28])).Bytes()
		if err2 != nil {
			if err2 == redis.Nil {
				return nil
			}
			return err2
		}

		kvPairs, oid2, _ := fastParseKMRVal(val)
		if kvPairs == 1 && len(val) == 16 { // Without origin keys.
			_, err2 = tx.HDel(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28])).Result()
			oid = oid2
			return err2
		}

		// need search the val by comparing key_origin in val.
		found, oid2, newVal := deleteKeyInVal(key, val, kvPairs)
		if !found {
			return nil
		}
		if len(newVal) == 0 {
			_, err2 = tx.HDel(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28])).Result()
		} else {
			_, err2 = tx.HSet(ctx, xstrconv.ToString(buf[:8]), xstrconv.ToString(buf[8:28]), newVal).Result()
		}
		oid = oid2
		return err2
	}, key)

	return oid, err
}

// getKeyHashes gets key_in_redis_hset & filed_name.
// https://github.com/zaibyte/zaifs/issues/58#issuecomment-977
func getKeyHashes(key string) (rKey uint64, field [20]byte) {
	return xxh3.HashString(key) % bktCnt, sha1.Sum(xstrconv.ToBytes(key))
}

// makeKMRVal makes val in redis hset with single oid, size pair.
// This function is designed for Insert, which means `p` should be started with key & field in hset,
// we should jump over.
func makeKMRVal(p []byte, oid uint64, size uint32) {

	off := 28

	binary.LittleEndian.PutUint32(p[off:off+4], 1) // cnt
	off += 4
	binary.LittleEndian.PutUint64(p[off:off+8], oid)
	off += 8
	binary.LittleEndian.PutUint32(p[off:off+4], size)
}

func (r *KomapRedis) txn(ctx context.Context, txf func(tx *redis.Tx) error, keys ...string) error {

	var err error

	l := &r.txlocks[xxh3.HashString(keys[0])%uint64(len(r.txlocks))]

	l.Lock()
	defer l.Unlock()

	// TODO: enable retry for some of idempodent transactions
	var retryOnFailture = false

	// There is no need to wait for a long time for the next trying.
	// redis is fast.
	retries := &orpc.Retryer{
		MinSleep: 5 * time.Microsecond,
		MaxTried: 0,
		MaxSleep: time.Second,
	}
	for i := 0; i < 50; i++ {
		err = r.RDB.Watch(ctx, txf, keys...)
		if shouldRetry(err, retryOnFailture) {
			retries.GetSleepDuration(i, int64(len(keys)))
			continue
		}
		return err
	}
	return err
}

type timeoutError interface {
	Timeout() bool
}

func shouldRetry(err error, retryOnFailure bool) bool {
	switch err {
	case redis.TxFailedErr:
		return true
	case io.EOF, io.ErrUnexpectedEOF:
		return retryOnFailure
	case nil, context.Canceled, context.DeadlineExceeded:
		return false
	}

	if v, ok := err.(timeoutError); ok && v.Timeout() {
		return retryOnFailure
	}

	s := err.Error()
	if s == "ERR max number of clients reached" {
		return true
	}
	ps := strings.SplitN(s, " ", 3)
	switch ps[0] {
	case "LOADING":
	case "READONLY":
	case "CLUSTERDOWN":
	case "TRYAGAIN":
	case "MOVED":
	case "ASK":
	case "ERR":
		if len(ps) > 1 {
			switch ps[1] {
			case "DISABLE":
				fallthrough
			case "NOWRITE":
				fallthrough
			case "NOREAD":
				return true
			}
		}
		return false
	default:
		return false
	}
	return true
}

// fastParseKMRVal parse value without the origin key(string).
func fastParseKMRVal(p []byte) (kvPairs uint32, oid uint64, size uint32) {

	kvPairs = binary.LittleEndian.Uint32(p[:4])
	if kvPairs != 1 {
		return
	}
	oid = binary.LittleEndian.Uint64(p[4:12])
	size = binary.LittleEndian.Uint32(p[12:16])
	return
}

// getOIDSizeByIdx gets oid & size by the index in the oid_pairs in val.
func getOIDSizeByIdx(p []byte, i int) (oid uint64, size uint32) {

	oid = binary.LittleEndian.Uint64(p[4+12*i : 4+12*i+8])
	size = binary.LittleEndian.Uint32(p[4+12*i+8 : 4+12*i+8+4])
	return
}

type kvPair struct {
	oid  uint64
	size uint32
	key  []byte
}

// insertKeyInVal inserts key in val(in redis_hset),
// return non-zero oid if found the same oid & it should be deleted (have been replaced by new one in newVal).
// return unchanged if nothing changed in val. (caused by found same key_origin & oid & size pair).
func insertKeyInVal(key string, val []byte, oid uint64, size uint32) (unchanged bool, oldOID uint64, newVal []byte) {

	kvPairs := binary.LittleEndian.Uint32(val[:4])

	// allocate the possible biggest buf first.
	newVal = make([]byte, len(val)+12+2+len(key)+2) // The collision rate is extremely low, it's okay no pool here.

	keyBytes := xstrconv.ToBytes(key)
	start := 4 + 12*kvPairs

	pairs := make([]kvPair, 0, kvPairs+1)

	idx := 0

	found := false

	for start < uint32(len(val)) {

		oid2, size2 := getOIDSizeByIdx(val, idx)
		kLen := uint32(binary.LittleEndian.Uint16(val[start : start+2]))
		key2 := val[start+2 : start+2+kLen]

		if bytes.Equal(keyBytes, key2) { // If found, replacing old one.

			if oid2 == oid && size2 == size {
				return true, 0, val // remain the same.
			}
			oldOID = oid2
			oid2 = oid // Using new value.
			size2 = size
			found = true // continue for collecting all pairs need to be reserved.
		}

		pairs = append(pairs, kvPair{
			oid:  oid2,
			size: size2,
			key:  key2,
		})

		start += uint32(kLen) + 2
		idx++
	}

	if len(pairs) == 0 { // Need to add first one which has no origin key.
		oid2, size2 := getOIDSizeByIdx(val, idx)

		pairs = append(pairs, kvPair{
			oid:  oid2,
			size: size2,
			key:  nil,
		})
	}

	newKvPairs := kvPairs
	if !found { // If not found, just adding new one.
		pairs = append(pairs, kvPair{
			oid:  oid,
			size: size,
			key:  keyBytes,
		})
		newKvPairs += 1
	}

	binary.LittleEndian.PutUint32(newVal[:4], newKvPairs)
	for i, p := range pairs {
		binary.LittleEndian.PutUint64(newVal[4+i*12:4+i*12+8], p.oid)
		binary.LittleEndian.PutUint32(newVal[4+i*12+8:4+i*12+8+4], p.size)
	}

	newStart := 4 + 12*newKvPairs
	for _, p := range pairs {
		binary.LittleEndian.PutUint16(newVal[newStart:newStart+2], uint16(len(p.key)))
		copy(newVal[newStart+2:], p.key)
		newStart += 2 + uint32(len(p.key))
	}
	return false, oldOID, newVal[:newStart]
}

// deleteKeyInVal deletes key in val(in redis_hset), return true if found.
func deleteKeyInVal(key string, val []byte, kvPairs uint32) (found bool, oid uint64, newVal []byte) {

	newVal = make([]byte, len(val)) // The collision rate is extremely low, it's okay no pool here.

	keyBytes := xstrconv.ToBytes(key)
	start := 4 + 12*kvPairs

	pairs := make([]kvPair, 0, kvPairs)

	idx := 0
	for start < uint32(len(val)) {

		oid2, size2 := getOIDSizeByIdx(val, idx)
		kLen := uint32(binary.LittleEndian.Uint16(val[start : start+2]))
		key2 := val[start+2 : start+2+kLen]

		if bytes.Equal(keyBytes, val[start+2:start+2+kLen]) {
			oid = oid2
			found = true // continue for collecting all pairs need to be reserved.
			start += uint32(kLen) + 2
			idx++
			continue
		}

		pairs = append(pairs, kvPair{
			oid:  oid2,
			size: size2,
			key:  key2,
		})

		start += uint32(kLen) + 2
		idx++
	}

	// If not found, must be first one if first keyLen is 0.
	foundNoOriginKey := false
	if !found {
		if len(pairs[0].key) == 0 {
			found = true
			foundNoOriginKey = true
			oid = pairs[0].oid
			pairs = pairs[1:]
		}
	}

	if found { // need newVal.
		if kvPairs-1 == 0 {
			return true, oid, nil
		}
		binary.LittleEndian.PutUint32(newVal[:4], kvPairs-1)

		for i, p := range pairs {
			binary.LittleEndian.PutUint64(newVal[4+i*12:4+i*12+8], p.oid)
			binary.LittleEndian.PutUint32(newVal[4+i*12+8:4+i*12+8+4], p.size)
		}

		if kvPairs-1 == 1 && !foundNoOriginKey {
			return true, oid, newVal[:16]
		}

		newStart := 4 + 12*(kvPairs-1)
		for _, p := range pairs {
			binary.LittleEndian.PutUint16(newVal[newStart:newStart+2], uint16(len(p.key)))
			copy(newVal[newStart+2:], p.key)
			newStart += 2 + uint32(len(p.key))
		}
		return true, oid, newVal[:newStart]
	}

	return false, 0, val
}

// searchKeyInVal searches key in val(in redis_hset), return true if found.
func searchKeyInVal(key string, val []byte, kvPairs uint32) (found bool, oid uint64, size uint32) {

	keyBytes := xstrconv.ToBytes(key)
	keyLen := uint32(len(key))
	start := 4 + 12*kvPairs

	idx := 0
	for start < uint32(len(val)) {

		kLen := uint32(binary.LittleEndian.Uint16(val[start : start+2]))
		if kLen == keyLen {

			if bytes.Equal(keyBytes, val[start+2:start+2+kLen]) {
				oid, size = getOIDSizeByIdx(val, idx)
				return true, oid, size
			}
		}

		start += uint32(kLen) + 2
		idx++
	}
	// If not found, must be first one if first keyLen is 0.
	if binary.LittleEndian.Uint16(val[4+12*kvPairs:]) == 0 {
		return true, binary.LittleEndian.Uint64(val[4:12]), binary.LittleEndian.Uint32(val[12:16])
	}

	return false, 0, 0
}
