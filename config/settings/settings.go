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

// Settings is the global settings of zai.
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
	ExtV1    uint16 = 1
	ExtVtest uint16 = 666
)

var ExtAvailVersion = []uint16{ExtV1, ExtVtest}

// Zai has three different isolation levels.
const (
	IsolationInstance = "instance"
	IsolationDisk     = "disk"
	IsolationNone     = "none"
)
