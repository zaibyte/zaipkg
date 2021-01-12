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

// package xhttp provides http implementation for Zai.
// Zai uses HTTP/1.1 for handling management requests,
// these requests are not performance sensitive,
// HTTP/1.1 is a good choice to make normal operations easier.
//
// Compare with HTTP/1.1 in standard lib:
// 1. Add some default server/client configs.
// 2. Add E2E checksum.
// 3. Add private headers.
// 4. Wrap some basic methods/functions, make it easier to use.
package xhttp

import (
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cast"
)

// These header names will be added in HTTP header in Zai.
const (
	ReqIDHeader    = "x-zai-request-id"
	ChecksumHeader = "x-zai-checksum"
)

func reqIDStrToInt(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}

// FillPath fills the julienschmidt/httprouter style path.
func FillPath(path string, kv map[string]interface{}) string {
	if kv == nil {
		return path
	}

	for k, v := range kv {
		vs := cast.ToString(v)
		path = strings.Replace(path, ":"+k, vs, 1)
	}
	return path
}

// ParsePath parses the julienschmidt/httprouter style path.
func ParsePath(p httprouter.Params, name string, val interface{}) {

	vs := p.ByName(name)
	if vs == "" {
		return
	}

	switch v := val.(type) {
	case *string:
		*v = vs
	case *int64:
		vint, err := strconv.ParseInt(vs, 10, 64)
		if err == nil {
			*v = vint
		}
	case *float64:
		vfloat, err := strconv.ParseFloat(vs, 64)
		if err == nil {
			*v = vfloat
		}
	case *uint16:
		*v = cast.ToUint16(vs)
	case *uint64:
		vuint, err := strconv.ParseUint(vs, 10, 64)
		if err == nil {
			*v = vuint
		}
	case *bool:
		vbool, err := strconv.ParseBool(vs)
		if err == nil {
			*v = vbool
		}
	}
}
