package uid

import (
	"strings"

	"github.com/google/uuid"
)

const DiskIDLen = 36

// IsValidDiskID returns the diskID is valid in Zai or not.
func IsValidDiskID(diskID string) bool {

	if len(diskID) != DiskIDLen {
		return false
	}

	if strings.ToLower(diskID) != diskID {
		return false
	}

	_, err := uuid.Parse(diskID)
	if err != nil {
		return false
	}

	return true
}
