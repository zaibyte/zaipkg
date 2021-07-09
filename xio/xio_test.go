package xio

import (
	"testing"
)

func TestIsReqRead(t *testing.T) {
	rrs := []uint64{
		ReqObjRead,
		ReqChunkRead,
		ReqGCRead}

	for _, r := range rrs {
		if !IsReqRead(r) {
			t.Fatal("mismatched")
		}
	}

	wrs := []uint64{
		ReqObjWrite,
		ReqChunkWrite,
		ReqGCWrite,
		ReqMetaWrite}

	for _, r := range wrs {
		if IsReqRead(r) {
			t.Fatal("mismatched")
		}
	}
}

func BenchmarkAsyncRequestChan(b *testing.B) {
	ch := make(chan *AsyncRequest, 4096)

	go func() {
		for {
			<-ch
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- new(AsyncRequest)
	}
}

func BenchmarkAsyncRequestChanMultiProducer(b *testing.B) {
	ch := make(chan *AsyncRequest, 4096)

	go func() {
		for {
			ar := <-ch
			ReleaseAsyncRequest(ar)
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			ch <- AcquireAsyncRequest()
		}
	})
}

func BenchmarkAsyncRequestChanMultiProducerTwoChan(b *testing.B) {
	ch := make(chan *AsyncRequest, 4096)
	ch2 := make(chan *AsyncRequest, 4096)

	go func() {
		for {
			select {
			case ar := <-ch:
				ReleaseAsyncRequest(ar)
			default:
				ar := <-ch2
				ReleaseAsyncRequest(ar)
			}
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			if i&1 == 1 {
				ch <- AcquireAsyncRequest()
			} else {
				ch2 <- AcquireAsyncRequest()
			}
		}
	})
}
