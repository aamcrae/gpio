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
	"unsafe"
)

const i2cCode uintptr = 7

var (
	i2cRetries       = iocW(i2cCode, 1, unsafe.Sizeof(uint32(0)))
	i2cTimeout       = iocW(i2cCode, 2, unsafe.Sizeof(uint32(0)))
	i2cSlave         = iocW(i2cCode, 3, unsafe.Sizeof(uint32(0)))
	i2cTenBit        = iocW(i2cCode, 4, unsafe.Sizeof(uint32(0)))
	i2cFuncs         = iocW(i2cCode, 5, unsafe.Sizeof(uint32(0)))
	i2cSlaveForce    = iocW(i2cCode, 6, unsafe.Sizeof(uint32(0)))
	i2cRdWr          = iocW(spiCode, 7, unsafe.Sizeof(i2c_rdwr{}))
)

type I2C struct {
	bus  int
	file *os.File
}

type i2c_rdwr struct {
	msgs uintptr
	count uint32
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
	return i2, nil
}

// Close closes the device.
func (i2 *I2C) Close() {
	i2.file.Close()
}
