// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xhttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"g.tesamc.com/IT/zaipkg/orpc"

	"g.tesamc.com/IT/zaipkg/version"
	"g.tesamc.com/IT/zaipkg/xlog"
	_ "g.tesamc.com/IT/zaipkg/xlog/xlogtest"
)

var (
	testServer  *httptest.Server
	testSrvAddr string

	testClient *Client
)

func init() {

	srv := NewServer(&ServerConfig{
		IdleTimeout:       0,
		ReadHeaderTimeout: 0,
	})
	testServer = httptest.NewServer(srv.srv.Handler)
	testSrvAddr = testServer.URL
	testClient, _ = NewDefaultClient()
}

func TestServerChecksum(t *testing.T) {
	// 1. Has check (passing checksum header)
	_, err := testClient.Version(testSrvAddr, 0)
	if err != nil {
		t.Fatal(err)
	}

	// 2. No check (no checksum header)
	req, err := http.NewRequest(http.MethodGet,
		testSrvAddr+"/v1/code-version", nil)
	if err != nil {
		return
	}
	resp, err := testClient.c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("should ok")
	}
	resp.Body.Close()

	// 3. Pass wrong checksum
	req, err = http.NewRequest(http.MethodGet,
		testSrvAddr+"/v1/code-version", nil)
	if err != nil {
		return
	}
	req.Header.Set(ChecksumHeader, "1")
	resp, err = testClient.c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Fatal("should bad request")
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	// See ReplyError for more details.
	err = errors.New(string(buf[:len(buf)-1])) // drop \n
	if errors.Is(err, orpc.ErrChecksumMismatch) {
		t.Fatal("error mismatched")
	}
	defer resp.Body.Close()
}

func TestServerDebug(t *testing.T) {

	err := testClient.Debug(testSrvAddr, true, 0)
	if err != nil {
		t.Fatal(err)
	}

	if xlog.GetLvl() != "debug" {
		t.Fatal("debug on failed")
	}

	err = testClient.Debug(testSrvAddr, false, 0)
	if err != nil {
		t.Fatal(err)
	}

	if xlog.GetLvl() != "info" {
		t.Fatal("debug off failed")
	}
}

func TestServerVersion(t *testing.T) {

	ret, err := testClient.Version(testSrvAddr, 0)
	if err != nil {
		t.Fatal(err)
	}

	if ret.Version != version.ReleaseVersion {
		t.Fatal("ReleaseVersion mismatch")
	}
	if ret.GitBranch != version.GitBranch {
		t.Fatal("GitBranch mismatch")
	}
	if ret.GitHash != version.GitHash {
		t.Fatal("GitBranch mismatch")
	}
}

func TestFillPath(t *testing.T) {
	path := "/test/:k0/:k1/:k2"
	kv := make(map[string]string)
	kv["k0"] = "v0"
	kv["k1"] = "v1"
	kv["k2"] = "v2"

	act := FillPath(path, kv)
	exp := "/test/v0/v1/v2"
	if act != exp {
		t.Fatal("mismatch")
	}
}
