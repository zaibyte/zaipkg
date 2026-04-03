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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/zaibyte/zaipkg/uid"

	"github.com/zaibyte/zaipkg/config"
	"github.com/zaibyte/zaipkg/version"
	"github.com/zaibyte/zaipkg/xchecksum"
	"github.com/zaibyte/zaipkg/xlog"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni/v2"
)

// ServerConfig is the config of Server.
type ServerConfig struct {
	Address string

	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

const (
	defaultIdleTimeout       = 75 * time.Second
	defaultReadHeaderTimeout = 3 * time.Second
)

// Server implements methods to build & run a HTTP server.
type Server struct {
	cfg    *ServerConfig
	router *httprouter.Router
	SVR    *http.Server

	middle *negroni.Negroni
}

func parseConfig(cfg *ServerConfig) {

	config.Adjust(&cfg.IdleTimeout, defaultIdleTimeout)
	config.Adjust(&cfg.ReadHeaderTimeout, defaultReadHeaderTimeout)
}

// NewServer creates a Server.
//
// Warn: Be sure you have run InitGlobalLogger before call it.
func NewServer(cfg *ServerConfig) (s *Server) {

	parseConfig(cfg)

	s = &Server{
		cfg: cfg,
	}

	s.addDefaultHandler()

	s.SVR = &http.Server{
		Addr:     cfg.Address,
		ErrorLog: log.New(xlog.GetLogger(), "", 0),
		Handler:  s.router,

		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	s.RegisterDefaultMiddleware()

	return
}

// AddHandler helps to add handler to Server.
func (s *Server) AddHandler(method, path string, handler httprouter.Handle) {

	s.router.Handle(method, path, handler)
}

// RegisterDefaultMiddleware registers default middleware and replacing origin handler.
func (s *Server) RegisterDefaultMiddleware() {
	s.middle = negroni.New(new(withRecovery), new(withCheck), new(withReqid))
	s.middle.UseHandler(s.router)
	s.SVR.Handler = s.middle
}

// GetHandler gets Server's http.Handler.
func (s *Server) GetHandler() http.Handler {
	return s.SVR.Handler
}

// Start starts the Server.
func (s *Server) Start() {

	go func() {
		_ = s.SVR.ListenAndServe()
	}()
}

// Close closes Server.
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_ = s.SVR.Shutdown(ctx)
}

// --- Default Handler ---- //

// addDefaultHandler add default handler.
func (s *Server) addDefaultHandler() {
	if s.router == nil {
		s.router = httprouter.New()
	}

	s.AddHandler(http.MethodPut, "/v1/debug-log/:cmd", s.debug)
	s.AddHandler(http.MethodGet, "/v1/code-version", s.version)
}

func (s *Server) debug(w http.ResponseWriter, _ *http.Request,
	p httprouter.Params) {

	reqIDS := w.Header().Get(ReqIDHeader)
	reqID := reqIDStrToInt(reqIDS)

	cmd := p.ByName("cmd")
	switch cmd {
	case "on":
		_ = xlog.SetLevel("debug")
		xlog.DebugID(reqID, "debug on")
	default:
		_ = xlog.SetLevel("info")
		xlog.InfoID(reqID, "debug off")
	}

	ReplyCode(w, http.StatusOK)
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	ReplyJson(w, &version.Info{
		Version:   version.ReleaseVersion,
		GitHash:   version.GitHash,
		GitBranch: version.GitBranch,
	}, http.StatusOK)
}

// Reply replies HTTP request.
//
// Usage:
// As return function in http Handler.
//
// Warn:
// Be sure you have called xlog.InitGlobalLogger.
// If any wrong in the write resp process, it would be written into the log.

// ReplyCode replies to the request with the empty message and HTTP code.
func ReplyCode(w http.ResponseWriter, statusCode int) {

	ReplyJson(w, nil, statusCode) // Only reply code, no need check resp.body.
}

// ReplyError replies to the request with the specified error message and HTTP code.
func ReplyError(w http.ResponseWriter, msg string, statusCode int) {

	if msg == "" {
		msg = http.StatusText(statusCode)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)

	_, err := fmt.Fprintln(w, msg)
	if err != nil {
		xlog.ErrorID(reqIDStrToInt(w.Header().Get(ReqIDHeader)), makeReplyErrMsg(err))
	}
}

// ReplyJson replies to the request with specified ret(in JSON) and HTTP code.
func ReplyJson(w http.ResponseWriter, ret interface{}, statusCode int) {

	var msg []byte
	if ret != nil {
		msg, _ = json.Marshal(ret)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
	w.WriteHeader(statusCode)

	w.Header().Set(ChecksumHeader, strconv.FormatInt(int64(xchecksum.Sum32(msg)), 10))

	_, err := w.Write(msg)
	if err != nil {
		xlog.ErrorID(reqIDStrToInt(w.Header().Get(ReqIDHeader)), makeReplyErrMsg(err))
	}
}

// GetReqID gets request id from request.
func GetReqID(req *http.Request) uint64 {
	reqid := reqIDStrToInt(req.Header.Get(ReqIDHeader))
	if reqid == 0 {
		return uid.MakeReqID()
	}
	return reqid
}

func makeReplyErrMsg(err error) string {
	return fmt.Sprintf("write resp failed: %s", err.Error())
}
