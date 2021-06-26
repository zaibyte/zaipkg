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

// Package settings is the global settings of zai.
// Don't modify it unless you totally know what will happen.
package settings

const (
	// DefaultLogRoot is the default log files path root.
	// e.g.:
	// <DefaultLogRoot>/<appName>/access.log
	// & <DefaultLogRoot>/<appName>/error.log
	DefaultLogRoot = "/var/log/zai"
)

const (
	MaxObjectSize = 4 * 1024 * 1024 // 4MiB.
)

const (
	ExtV1 uint16 = 1
)

var ValidExtVersions = []uint16{ExtV1}

// Zai has three different isolation levels.
const (
	IsolationInstance = "instance"
	IsolationDisk     = "disk"
	IsolationNone     = "none"
)

var ValidIsolationLevels = []string{IsolationInstance, IsolationDisk, IsolationNone}

// DefaultIsolationLevel is IsolationInstance, enough for giving enough protection:
// 1. Each machine in the same box will only be placed in the same IDC. (see arch docs for details)
// 2. Instance isolation is enough storage for giving high durability.
const DefaultIsolationLevel = IsolationInstance

// DefaultReplicas is 2.
// In Tesamc, we will start at 2 replicas first for saving overhead.
// There are other replicas in public/private cloud storage. 2 replicas just for speeding up repairing.
const DefaultReplicas = 2

const (
	kb = 1024
	mb = 1024 * kb
	gb = 1024 * mb
)

// DefaultExtV1SegSize is 1GB, which means the extent size is 256GB.
// For a 8TB NVMe driver(raw capacity), in real world, there will be space for over-provisioning & other things.
// So we have about less than 30 extents on each disk.
//
// It's obvious that the bigger extent, the lower rate of losing group when there are broken disks.
// But we can't make it too bigger either, because we may lose the property of distributed repairing,
// we hope if there is a broken disk, more disks could help to reconstruct the data, it'll reduce the
// load of reconstruction on disks in avg. and speeding up the process.
//
// More details about the rate of group failed see: https://g.tesamc.com/IT/zai-docs/issues/19
const DefaultExtV1SegSize = gb
