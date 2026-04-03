package xhttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"

	"github.com/zaibyte/zaipkg/orpc"
	"github.com/zaibyte/zaipkg/uid"
	"github.com/zaibyte/zaipkg/xchecksum"
	"github.com/zaibyte/zaipkg/xlog"
)

type withRecovery struct{}

func (p *withRecovery) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			xlog.Error(fmt.Sprintf("panic occured: %v\nStack trace: %s", x, stackTrace[:n]))
		}
	}()

	next(w, r)
}

type withReqid struct{}

func (p *withReqid) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqID := r.Header.Get(ReqIDHeader)
	if reqID == "" {
		reqID = strconv.FormatUint(uid.MakeReqID(), 10)
	}
	w.Header().Set(ReqIDHeader, reqID)
	next(w, r)
}

type withCheck struct{}

func (p *withCheck) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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
		incoming, err2 := strconv.Atoi(clientSumStr)
		if err2 != nil {
			ReplyError(w, orpc.ErrBadRequest.Error(), http.StatusBadRequest)
			return
		}
		h := xchecksum.New()
		_, _ = h.Write([]byte(r.URL.RequestURI()))
		_, _ = h.Write(b)
		act := int(h.Sum32())
		if incoming != act {
			ReplyError(w, orpc.ErrChecksumMismatch.Error(), http.StatusBadRequest)
			return
		}
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	next(w, r)
}
