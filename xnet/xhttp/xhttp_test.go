package xhttp

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestFillPath(t *testing.T) {
	path := "/test/:k0/:k1_1/:k2/:k3/:k4"
	kv := make(map[string]interface{})
	kv["k0"] = "v0"
	kv["k1_1"] = int64(1)
	kv["k2"] = float64(0.2)
	kv["k3"] = uint64(3)
	kv["k4"] = true

	pAfterFill := FillPath(path, kv)

	router := httprouter.New()

	routed := false
	router.Handle("GET", path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		routed = true

		for k, v := range kv {
			switch v.(type) {
			case string:
				var act string
				ParsePath(ps, k, &act)
				assert.Equal(t, v, act)
			case int64:
				var act int64
				ParsePath(ps, k, &act)
				assert.Equal(t, v, act)
			case float64:
				var act float64
				ParsePath(ps, k, &act)
				assert.Equal(t, v, act)
			case uint64:
				var act uint64
				ParsePath(ps, k, &act)
				assert.Equal(t, v, act)
			}
		}
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest("GET", pAfterFill, nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}

	var ps httprouter.Params
	ps = make([]httprouter.Param, len(kv))
	i := 0
	for k, v := range kv {
		ps[i] = httprouter.Param{k, cast.ToString(v)}
	}

}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}
