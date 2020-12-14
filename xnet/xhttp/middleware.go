package xhttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xchecksum"
	"g.tesamc.com/IT/zaipkg/xlog"

	"github.com/julienschmidt/httprouter"
)

func withRecovery(next httprouter.Handle) httprouter.Handle {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			xlog.Error(fmt.Sprintf("panic occured: %v\nStack trace: %s", x, stackTrace[:n]))
		}
	}()

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		next(w, r, p)
	}
}

func withReqid(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		reqID := r.Header.Get(ReqIDHeader)
		if reqID == "" {
			reqID = strconv.FormatUint(uid.MakeReqID(), 10)
		}
		w.Header().Set(ReqIDHeader, reqID)
		next(w, r, p)
	}
}

func withCheck(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		enableCheck := true
		clientSumStr := r.Header.Get(ChecksumHeader)
		if clientSumStr == "" {
			enableCheck = false
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ReplyCode(w, http.StatusInternalServerError)
			return
		}

		if enableCheck {
			incoming, err := strconv.Atoi(clientSumStr)
			if err != nil {
				ReplyError(w, orpc.ErrBadRequest.Error(), http.StatusBadRequest)
				return
			}
			h := xchecksum.New()
			h.Write([]byte(r.URL.RequestURI()))
			h.Write(b)
			act := int(h.Sum32())
			if incoming != act {
				ReplyError(w, orpc.ErrChecksumMismatch.Error(), http.StatusBadRequest)
				return
			}
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(b))

		next(w, r, p)
	}
}
