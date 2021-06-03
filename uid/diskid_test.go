package uid

import (
	"strings"
	"testing"

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
