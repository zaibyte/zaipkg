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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"g.tesamc.com/IT/zaipkg/orpc"

	"g.tesamc.com/IT/zaipkg/xchecksum"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/version"
)

// Client is an xhttp client.
type Client struct {
	c         *http.Client
	addScheme func(url string) string
}

const (
	defaultDialTimeout         = 3000 * time.Millisecond
	defaultRespTimeout         = 16 * time.Second
	defaultKeepAlive           = 75 * time.Second
	defaultMaxIdleConns        = 100
	defaultMaxIdleConnsPerHost = 10
	defaultIdleConnTimeout     = 90 * time.Second
)

var (
	// DefaultTransport is a HTTP/1.1 transport.
	DefaultTransport = &http.Transport{
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		IdleConnTimeout:       defaultIdleConnTimeout,
		ResponseHeaderTimeout: defaultRespTimeout,
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
	}
)

// NewDefaultClient creates a Client with default configs.
func NewDefaultClient() (*Client, error) {
	return NewClientWithTransport(DefaultTransport), nil
}

// NewClientWithTransport creates a Client with a transport.
// If transport == nil, use DefaultTransport.
func NewClientWithTransport(transport *http.Transport) *Client {

	if transport == nil {
		transport = DefaultTransport
	}

	return &Client{
		c: &http.Client{
			Transport: transport,
		},
		addScheme: addHTTPScheme,
	}
}

func addHTTPScheme(url string) string {
	return addScheme(url, "http://")
}

// addScheme adds HTTP scheme if need.
func addScheme(url string, scheme string) string {
	if !strings.HasPrefix(url, scheme) {
		url = scheme + url
	}
	return url
}

// Do sends an HTTP request and returns an HTTP response.
//
// A >= 400 status code DO cause an error.
// All >= 400 response will be closed.
//
// On error, any Response can be ignored.
func (c *Client) Request(ctx context.Context, method, url string, reqID uint64, buf []byte) (resp *http.Response, err error) {

	url = c.addScheme(url)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(buf))
	if err != nil {
		return
	}

	var reqidStr string
	if reqID == 0 {
		reqidStr = strconv.FormatUint(uid.MakeReqID(), 10)
	}
	req.Header.Set(ReqIDHeader, reqidStr)

	h := xchecksum.New()
	_, _ = h.Write([]byte(req.URL.RequestURI()))
	_, _ = h.Write(buf)
	req.Header.Set(ChecksumHeader, strconv.Itoa(int(h.Sum32())))

	resp, err = c.c.Do(req)
	if err != nil {
		return
	}

	ch := resp.Header.Get(ChecksumHeader)
	if ch != "" {
		incoming, err := strconv.Atoi(ch)
		if err != nil {
			io.Copy(ioutil.Discard, resp.Body)
			return resp, orpc.ErrBadRequest
		}

		b, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return resp, orpc.ErrInternalServer
		}

		if incoming != int(xchecksum.Sum32(b)) {
			return resp, orpc.ErrChecksumMismatch
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	}

	if resp.StatusCode/100 >= 4 {

		err = errors.New(http.StatusText(resp.StatusCode))

		if resp.ContentLength > 0 && method != http.MethodHead {
			buf, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return resp, orpc.ErrInternalServer
			}
			// See ReplyError for more details.
			errMsg := string(buf[:len(buf)-1]) // drop \n
			err = orpc.StrError(errMsg)
		}
		io.Copy(ioutil.Discard, resp.Body)
		return
	}

	return
}

// --- Default API ---- //
// --- All HTTP Servers in zai will have these APIs ---- //

const defaultTimeout = 3 * time.Second

// Debug opens/closes a server logger's debug level.
func (c *Client) Debug(addr string, on bool, reqID uint64) (err error) {

	cmd := "off"
	if on {
		cmd = "on"
	}

	url := addr + "/v1/debug-log/" + cmd
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.Request(ctx, http.MethodPut, url, reqID, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return nil
}

// Version returns the code version of a server.
func (c *Client) Version(addr string, reqID uint64) (ver version.Info, err error) {

	url := addr + "/v1/code-version"
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.Request(ctx, http.MethodGet, url, reqID, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&ver)
	return
}
