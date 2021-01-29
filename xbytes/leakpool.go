package xbytes

// LeakyPool is a pool of bytes slice which will never be GC.
type LeakyPool struct {
	makeFn   func() []byte
	freeList chan []byte
}

// NewLeakyBuf creates a leaky buffer which can hold at most n buffer.
func NewLeakyBuf(n int, makeFn func() []byte) *LeakyPool {
	return &LeakyPool{
		makeFn:   makeFn,
		freeList: make(chan []byte, n),
	}
}

// Get returns a buffer from the leaky buffer or create a new buffer.
func (lb *LeakyPool) Get() (b []byte) {
	select {
	case b = <-lb.freeList:
	default:
		b = lb.makeFn()
	}
	return
}

// Put add the buffer into the free buffer pool for reuse.
func (lb *LeakyPool) Put(b []byte) {

	select {
	case lb.freeList <- b:
	default:
	}
	return
}
