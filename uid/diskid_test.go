package uid

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jaypipes/ghw/pkg/block"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsValidDiskID(t *testing.T) {

	cases := []struct {
		diskID string
		exp    bool
	}{
		{
			uuid.NewString(),
			true,
		},
		{
			strings.ToUpper(uuid.NewString()),
			false,
		},
		{
			"1",
			false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.exp, IsValidDiskID(c.diskID))
	}
}

func TestGetSerial(t *testing.T) {

	blk, err := block.New()
	if err != nil {
		t.Fatal(err)
	}

	if len(blk.Disks) == 0 {
		t.Skip("zero")
	}

	for _, d := range blk.Disks {
		fmt.Printf("%#v\n", d)
		if len(d.Partitions) != 0 {
			fmt.Println(d.Partitions[0].UUID)
		}
	}
}
