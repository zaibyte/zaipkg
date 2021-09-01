package vfs

// SyncDir syncs directory.
// It will return invalid handle on Windows, so just return nil.
func SyncDir(fs FS, dir string) error {

	return nil
}
