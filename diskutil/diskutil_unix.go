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

// +build !windows

package diskutil

import "golang.org/x/sys/unix"

func getUsage(path string) (UsageState, error) {
	stat := unix.Statfs_t{}
	if err := unix.Statfs(path, &stat); err != nil {
		return UsageState{}, err
	}
	return UsageState{
		Size: uint64(stat.Bsize) * stat.Blocks,
		Free: uint64(stat.Bsize) * stat.Bfree, // Free in filesystem.
		Used: uint64(stat.Bsize) * (stat.Blocks - stat.Bfree),
	}, nil
}
