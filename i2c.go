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
	"fmt"
	"os"
	"time"
	"unsafe"
)

type I2cMsg struct {
	Addr  uint16
	Flags int
	Buf   []byte
}

const I2cMaxMsgs = 42

// Flags
const (
	I2cFlagRead   = 1 << iota // Message is to be read, not written
	I2cFlagTenBit             // Address is 10 bit address
)

const i2cCode uintptr = 7

var (
	i2cRetries    = iocW(i2cCode, 1, unsafe.Sizeof(uintptr(0)))
	i2cTimeout    = iocW(i2cCode, 2, unsafe.Sizeof(uintptr(0)))
	i2cSlave      = iocW(i2cCode, 3, unsafe.Sizeof(uintptr(0)))
	i2cTenBit     = iocW(i2cCode, 4, unsafe.Sizeof(uintptr(0)))
	i2cFuncs      = iocW(i2cCode, 5, unsafe.Sizeof(uintptr(0)))
	i2cSlaveForce = iocW(i2cCode, 6, unsafe.Sizeof(uintptr(0)))
	i2cRdWr       = iocW(spiCode, 7, unsafe.Sizeof(i2c_rdwr{}))
)

type I2C struct {
	bus   int
	file  *os.File
	funcs uint32
}

type i2c_rdwr struct {
	msgs  uintptr
	count uint32
}

type i2c_msg struct {
	addr  uint16
	flags uint16
	len   uint16
	_     uint16 // Explicit pad - the ioctl does not have this, caveat.
	buf   uintptr
}

// NewI2C creates and initialises a new I2C device.
func NewI2C(bus int) (*I2C, error) {
	i2 := new(I2C)
	i2.bus = bus
	var err error
	i2.file, err = os.OpenFile(fmt.Sprintf("/dev/i2c-%d", i2.bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	var funcs uintptr
	err = ioctl(i2.file.Fd(), i2cFuncs, uintptr(unsafe.Pointer(&funcs)))
	if err != nil {
		i2.Close()
		return nil, err
	}
	i2.funcs = uint32(funcs)
	i2.Timeout(time.Millisecond * 50)
	i2.Retries(3)
	return i2, nil
}

// Close closes the device.
func (i2 *I2C) Close() {
	i2.file.Close()
}

func (i2 *I2C) Timeout(tout time.Duration) error {
	// Round up to nearest 10 ms
	v := uintptr((tout.Milliseconds() + 9) / 10)
	return ioctl(i2.file.Fd(), i2cTimeout, v)
}

func (i2 *I2C) TenBit(ten bool) error {
	var v uintptr
	if ten {
		v = 1
	}
	return ioctl(i2.file.Fd(), i2cTenBit, v)
}

func (i2 *I2C) Retries(r int) error {
	return ioctl(i2.file.Fd(), i2cRetries, uintptr(r))
}

func (i2 *I2C) Message(msgs []I2cMsg) error {
	if len(msgs) == 0 || len(msgs) > I2cMaxMsgs {
		return os.ErrInvalid
	}
	m := make([]i2c_msg, len(msgs))
	mi := &i2c_rdwr{uintptr(unsafe.Pointer(&m[0])), uint32(len(m))}
	err := ioctl(i2.file.Fd(), i2cRdWr, uintptr(unsafe.Pointer(mi)))
	if err != nil {
		return err
	}
	// Walk through the buffers and adjust the lengths.
	return nil
}
