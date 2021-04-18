package metricutil

import (
	"fmt"
	"sort"
	"strings"
)

func CombNameLabel(name, label string) string {
	return fmt.Sprintf(`%s{%s}`, name, label)
}

// MakeMetricName makes github.com/VictoriaMetrics/metrics style metric name.
// e.g. `queue_size{queue="foobar",topic="baz"}`
func MakeMetricName(name string, labels map[string]string) string {
	l := MakeLabel(labels)
	return CombNameLabel(name, l)
}

func MakeLabel(labels map[string]string) string {
	ls := make([]string, 0, len(labels))
	for k, v := range labels {
		ls = append(ls, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(ls) // MakeMetricName will only run one time, the cost of sort is okay, but easier to read & test.
	return strings.Join(ls, ",")
}
