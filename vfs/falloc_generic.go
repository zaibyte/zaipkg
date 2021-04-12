// +build !linux,!darwin

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

package vfs

import "io"

func FAlloc(f File, length int64) error {
	df := f.(*DirectFile)

	curOff, err := df.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	size, err := df.Seek(length, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err = df.Seek(curOff, io.SeekStart); err != nil {
		return err
	}
	if length > size {
		return nil
	}
	return f.Truncate(length)
}
