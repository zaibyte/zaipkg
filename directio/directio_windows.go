// TODO Direct IO for windows

package directio

import (
	"os"
)

const (
	// Size to align the buffer to
	AlignSize = 4096

	// Minimum block size
	BlockSize = 4096
)

func OpenFile(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	return os.OpenFile(name, mode, perm)
}
