package diskutil

import (
	"fmt"
	"testing"
)

func TestGetDiskSN(t *testing.T) {
	fmt.Println(GetDiskSN("/log"))
}
