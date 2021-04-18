package metricutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeMetricName(t *testing.T) {
	labels := map[string]string{
		"a":   "a",
		"a_a": "a_a",
		"b":   "b",
		"b_b": "b_b",
		"c":   "/c/c",
	}
	name := "name"

	assert.Equal(t, `name{a="a",a_a="a_a",b="b",b_b="b_b",c="/c/c"}`, MakeMetricName(name, labels))
}
