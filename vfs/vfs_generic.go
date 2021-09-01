// +build !windows

package vfs

// SyncDir syncs directory.
func SyncDir(fs FS, dir string) error {

	f, err := fs.OpenDir(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Sync()
	return err
}
