package xtest

import (
	"flag"
)

var _propEnabled = flag.Bool("xtest.prop", false, "enable properties testing or not")

// IsPropEnabled returns enable properties testing or not.
func IsPropEnabled() bool {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *_propEnabled
}
