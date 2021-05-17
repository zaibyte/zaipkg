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

// Package diskutil implements methods to access disk status.
package diskutil

import (
	"errors"
	"strings"
	"syscall"

	"github.com/gyuho/linux-inspect/df"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// IsBroken returns an error is disk error or not.
// If the err is EIO or EROFS, the disk should be regard as broken.
// But EIO is more likely read bad sector, the driver firmware will remap the bad sector in the next writing,
// what's more, we have other monitor to detect disk broken, so EIO won't be signal of disk broken.
//
// The logic of EIO & EROFS is copied from Western Digital open source object storage:
// https://github.com/westerndigitalcorporation/blb/blob/master/internal/tractserver/manager.go
// func (m *Manager) toBlbError(err error) core.Error
//
// And further discussion: https://github.com/westerndigitalcorporation/blb/issues/8.
func IsBroken(err error) bool {
	if err == nil {
		return false
	}

	// EIO: I/O error
	// if errors.Is(err, syscall.EIO) {
	// 	return true
	// }

	// EROFS: Read-only file system, caused by
	// 1. VFS error,
	// 2. hard disk error
	if errors.Is(err, syscall.EROFS) {
		return true
	}

	return false
}

// UsageState Wraps Syscall Statfs.
type UsageState struct {
	Size uint64
	Free uint64
	Used uint64
}

// GetUsageState returns disk basic capacity state (unit: Byte).
func GetUsageState(path string) (UsageState, error) {
	return getUsage(path)
}

// GetFreeSize returns disk free space size (unit: Byte).
func GetFreeSize(path string) (free uint64, err error) {
	u, err := GetUsageState(path)
	if err != nil {
		return 0, err
	}
	return u.Free, nil
}

// GetDiskType gets disk device interface type by `df`.
//
// It regards all non-nvme devices as SATA in present.
func GetDiskType(path string) metapb.DiskType {
	rs, err := df.GetDefault(``)
	if err != nil {
		return metapb.DiskType_Disk_SATA
	}

	for _, r := range rs {
		if r.MountedOn == path {
			if strings.HasPrefix(r.FileSystem, "/dev/nvme") {
				return metapb.DiskType_Disk_NVMe
			}
		}
	}
	return metapb.DiskType_Disk_SATA
}
