package xhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xchecksum"
	"github.com/julienschmidt/httprouter"
)

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
