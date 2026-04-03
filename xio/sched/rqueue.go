package sched

import (
	"github.com/zaibyte/zaipkg/xio"

	"github.com/templexxx/tsc"
)

// ReqQueue is the xio.AsyncRequest queue.
type ReqQueue struct {
	queue chan *xio.AsyncRequest
}

func (p *ReqQueue) add(reqType uint64, f xio.File, offset int64, d []byte) (ar *xio.AsyncRequest, err error) {

	ar = xio.AcquireAsyncRequest()

	ar.Type = reqType
	ar.Data = d
	ar.File = f
	ar.Offset = offset
	ar.Err = make(chan error)
	ar.PTS = tsc.UnixNano()

	select {
	case p.queue <- ar:
	default:
		select {
		case p.queue <- ar:

		}
	}
	return ar, err
}
