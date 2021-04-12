// +build !windows

package vfs

import (
	"os"
	"path/filepath"
	"syscall"

	"g.tesamc.com/IT/zaipkg/directio"
	"github.com/templexxx/fnc"
)

// Create creates a new file read/write, and sync the directory.
// If failed, trying to remove dirty file.
func (directFS) Create(name string) (File, error) {
	f, err := directio.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC|syscall.O_CLOEXEC|fnc.O_NOATIME, 0666)
	if err != nil {
		return nil, err
	}
	err = SyncDir(DirectFS, filepath.Dir(name))
	if err != nil {
		_ = os.Remove(name)
		return nil, err
	}
	return &DirectFile{f}, nil
}
