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

// DoNothing does nothing, only for some framework testing to test pure framework cost.
// Using n to control the function total cost, actually a spin is inside this function.
//
// e.g. when n = 1, it'll cost 30-40ns.
// If you want 400ns wait, n could be 10.
//
// It's not a good idea to use time.Sleep as DoNothing, because it'll bring
// goroutine scheduler make cost unpredictable.
func DoNothing(n uint32) {
	spin(n)
}
