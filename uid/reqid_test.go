/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uid

import (
	"testing"
	"time"

	"github.com/templexxx/tsc"
)

func TestMakeParseReqID(t *testing.T) {

	// Because it's fast, second ts may not change.
	expTime := time.Unix(0, tsc.UnixNano())
	reqID := MakeReqID()
	ts := GetTSFromReqID(reqID)
	if expTime.Unix() != ts/int64(time.Second) &&
		expTime.Unix()+1 != ts/int64(time.Second) { // May meet critical point.
		t.Fatal("mismatch", expTime.Unix(), ts/int64(time.Second))
	}
}

func BenchmarkMakeReqID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeReqID()
	}
}
