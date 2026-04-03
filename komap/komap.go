package komap

import "context"

// KoMap is the key:oid mapping.
type KoMap interface {
	// Insert inserts key:oid_size pair.
	//
	// oldOID is the oid with the same key,
	// Insert will insert new one even if there is existed key.
	//
	// If the oldOID is different with the new one, caller has the responsibility to delete old on in Zai.
	// If there is no oldOID, return 0.
	Insert(ctx context.Context, key string, oid uint64, size uint32) (oldOID uint64, err error)
	Search(ctx context.Context, key string) (oid uint64, size uint32, err error)
	// Delete deletes key.
	// Returns oid for Zai.Client deletes it.
	// If not found, return nil.
	Delete(ctx context.Context, key string) (oid uint64, err error)
}
