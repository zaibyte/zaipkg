// Package xruntime provides helper tools for Go Runtime.
package xruntime

import "runtime"

// AutoGOMAXPROCS sets the maximum number of CPUs that can be executing
// simultaneously automatically.
//
// The number of logical CPUs on the local machine can be queried with NumCPU.
// This call will go away when the scheduler improves.
func AutoGOMAXPROCS() {

	p := runtime.NumCPU()
	// Actually Go can't handle multi-cores well enough,
	// in ByteDance(A large internet Company which uses Go heavily) Graph Database team,
	// one Go process will only have 20 cores at most.
	if p > 32 {
		p = 32
	}
	runtime.GOMAXPROCS(p)
}
