package xatomic

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
	"unsafe"

	"g.tesamc.com/IT/zaipkg/xbytes"
)

var (
	magic128AVX = xbytes.MakeAlignedBlock(16, 16)
)

func init() {
	binary.LittleEndian.PutUint64(magic128AVX[:8], 0xdeddeadbeefbeef)
	binary.LittleEndian.PutUint64(magic128AVX[8:], 0xdeddeadbeefbeef)
}

func TestAtomicLoad16BAVX(t *testing.T) {
	var x struct {
		before []uint8
		i      []byte
		after  []uint8
	}
	x.before = magic128AVX
	x.after = magic128AVX
	x.i = xbytes.MakeAlignedBlock(16, 16)

	for delta := uint64(1); delta+delta > delta; delta += delta {
		k := xbytes.MakeAlignedBlock(16, 16)
		AvxLoad16B(&x.i[0], &k[0])
		if !bytes.Equal(k[:], x.i) {
			t.Fatalf("delta=%d i=%d k=%d", delta, x.i, k)
		}

		xi0 := binary.LittleEndian.Uint64(x.i[:8])
		xi0 += delta
		xi1 := binary.LittleEndian.Uint64(x.i[8:])
		xi1 -= delta

		binary.LittleEndian.PutUint64(x.i[:8], delta+2)
		binary.LittleEndian.PutUint64(x.i[8:], ^(delta + 2))
	}
	if !bytes.Equal(x.before, magic128) || !bytes.Equal(x.after, magic128) {
		t.Fatal("wrong magic")
	}
}

func TestAtomicStore16BAVX(t *testing.T) {
	var x struct {
		before []uint8
		i      []byte
		after  []uint8
	}
	x.before = magic128
	x.after = magic128
	x.i = xbytes.MakeAlignedBlock(16, 16)

	v := xbytes.MakeAlignedBlock(16, 16)
	for delta := uint64(1); delta+delta > delta; delta += delta {
		AvxStore16B(&x.i[0], &v[0])
		if !bytes.Equal(v[:], x.i) {
			t.Fatalf("delta=%d i=%d", delta, x.i)
		}

		xi0 := binary.LittleEndian.Uint64(v[:8])
		xi0 += delta
		xi1 := binary.LittleEndian.Uint64(v[8:])
		xi1 -= delta

		binary.LittleEndian.PutUint64(v[:8], delta+2)
		binary.LittleEndian.PutUint64(v[8:], ^(delta + 2))
	}
	if !bytes.Equal(x.before, magic128) || !bytes.Equal(x.after, magic128) {
		t.Fatal("wrong magic")
	}
}

func hammerStoreLoadUint128AVX(t *testing.T, paddr unsafe.Pointer) {
	addr := (*byte)(paddr)
	v := xbytes.MakeAlignedBlock(16, 16)
	AvxLoad16B(addr, &v[0])
	v0 := binary.LittleEndian.Uint64(v[:8])
	v1 := binary.LittleEndian.Uint64(v[8:])

	if v0 != v1 {
		t.Fatalf("AVXUint128: %#x != %#x", v0, v1)
	}
	newV := xbytes.MakeAlignedBlock(16, 16)
	binary.LittleEndian.PutUint64(newV[:8], v0+1)
	binary.LittleEndian.PutUint64(newV[8:], v1+1)

	AvxStore16B(addr, &newV[0])
}

func TestHammerStoreLoadAVX(t *testing.T) {
	var tests []func(*testing.T, unsafe.Pointer)
	tests = append(tests, hammerStoreLoadUint128AVX)
	n := int(1e6)
	if testing.Short() {
		n = int(1e4)
	}
	const procs = 8
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(procs))
	for _, tt := range tests {
		c := make(chan int)
		val := xbytes.MakeAlignedBlock(16, 16)
		for p := 0; p < procs; p++ {
			go func() {
				for i := 0; i < n; i++ {
					tt(t, unsafe.Pointer(&val[0]))
				}
				c <- 1
			}()
		}
		for p := 0; p < procs; p++ {
			<-c
		}
	}
}

func TestNilDerefAVX(t *testing.T) {
	funcs := [...]func(){
		func() {
			var a, b [16]byte
			AtomicCAS16B(nil, &(a[0]), &(b[0]))
		},
		func() { AtomicLoad16B(nil) },
		func() { AtomicStore16B(nil, [16]byte{0}) },
	}
	for _, f := range funcs {
		func() {
			defer func() {
				runtime.GC()
				recover()
			}()
			f()
		}()
	}
}
