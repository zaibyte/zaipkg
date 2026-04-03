package xtest

import (
	"flag"
)

var _propEnabled = flag.Bool("xtest.prop", false, "enable properties testing or not")

// IsPropEnabled returns enable properties testing or not.
// Default is false.
//
// e.g.
// no properties testing: go test -xtest.prop=false -v or go test -v
// run properties testing: go test -xtest.prop=true -v
func IsPropEnabled() bool {
	if !flag.Parsed() {
		flag.Parse()
	}

	return *_propEnabled
}
