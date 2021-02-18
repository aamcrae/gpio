// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctl8(fd, req uintptr, b *byte) error {
	return ioctl(fd, req, uintptr(unsafe.Pointer(b)))
}

func ioctl32(fd, req uintptr, i *uint32) error {
	return ioctl(fd, req, uintptr(unsafe.Pointer(i)))
}

func ioctl(fd, req, arg uintptr) error {
	_, _, ep := unix.Syscall(unix.SYS_IOCTL, fd, req, arg)
	if ep != 0 {
		return ep
	}
	return nil
}

func iocR(code, nr, sz uintptr) uintptr {
	return iocreq(2, code, nr, sz)
}

func iocW(code, nr, sz uintptr) uintptr {
	return iocreq(1, code, nr, sz)
}

func iocRW(code, nr, sz uintptr) uintptr {
	return iocreq(1|2, code, nr, sz)
}

func iocreq(dir, code, nr, sz uintptr) uintptr {
	return (dir << 30) | (sz << 16) | (code << 8) | nr
}
