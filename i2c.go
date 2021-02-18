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

// I2cMsg represents a single read or write message.
type I2cMsg struct {
	Addr  uint16
	Flags int
	Buf   []byte
}

const I2cMaxMsgs = 42 // Maximum number of messages allowed in transaction

// Flags
const (
	I2cFlagRead   = 1 << iota // Message is to be read, not written
	I2cFlagTenBit             // Address is 10 bit address
)

const i2cCode uintptr = 7

var (
	i2cRetries    uintptr = 0x0701
	i2cTimeout    uintptr = 0x0702
	i2cSlave      uintptr = 0x0703
	i2cTenBit     uintptr = 0x0704
	i2cFuncs      uintptr = 0x0705
	i2cSlaveForce uintptr = 0x0706
	i2cRdWr       uintptr = 0x0707
)

// I2C represents one I2C device
type I2C struct {
	bus   int
	file  *os.File
	addr  uint16 // Default address
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

// Addr sets the default address.
func (i2 *I2C) Addr(addr uint16) error {
	if (i2.funcs & 0x0002) != 0 { // 10 bit address allowed
		if addr >= (1 << 10) {
			return os.ErrInvalid
		}
	} else if addr >= (1 << 7) {
		return os.ErrInvalid
	}
	i2.addr = addr
	return nil
}

// Timeout sets the default timeout for the bus.
func (i2 *I2C) Timeout(tout time.Duration) error {
	// Round up to nearest 10 ms
	v := uintptr((tout.Milliseconds() + 9) / 10)
	return ioctl(i2.file.Fd(), i2cTimeout, v)
}

// TenBit enables 10 bit addresses.
func (i2 *I2C) TenBit(ten bool) error {
	var v uintptr
	if ten {
		v = 1
	}
	return ioctl(i2.file.Fd(), i2cTenBit, v)
}

// Retries sets the default number of message retries.
func (i2 *I2C) Retries(r int) error {
	return ioctl(i2.file.Fd(), i2cRetries, uintptr(r))
}

// Read builds a message slice that writes an 8 bit register value to the
// device and then reads data from the peripheral device.
func (i2 *I2C) Read(reg byte, b []byte) error {
	m := make([]I2cMsg, 2)
	m[0].Addr = i2.addr
	m[0].Buf = []byte{reg}
	m[1].Addr = i2.addr
	m[1].Flags = I2cFlagRead
	m[1].Buf = b
	return i2.Message(m)
}

// Write builds a message to write a register address to the peripheral device
// followed by the byte data.
func (i2 *I2C) Write(reg byte, data []byte) error {
	m := make([]I2cMsg, 1)
	m[0].Addr = i2.addr
	m[0].Buf = append([]byte{reg}, data...)
	return i2.Message(m)
}

// ReadReg reads one 8 bit register from the peripheral device by
// writing a register address and then reading 1 byte from the device.
func (i2 *I2C) ReadReg(reg byte) (byte, error) {
	b := []byte{0}
	err := i2.Read(reg, b)
	return b[0], err
}

// WriteReg writes one register in the peripheral device.
func (i2 *I2C) WriteReg(reg, data byte) error {
	return i2.Write(reg, []byte{data})
}

// Message writes or reads the list of messages to/from the
// peripheral device.
func (i2 *I2C) Message(msgs []I2cMsg) error {
	if len(msgs) == 0 || len(msgs) > I2cMaxMsgs {
		return os.ErrInvalid
	}
	m := make([]i2c_msg, len(msgs))
	mi := &i2c_rdwr{uintptr(unsafe.Pointer(&m[0])), uint32(len(m))}
	for i := range msgs {
		m[i].addr = msgs[i].Addr
		m[i].len = uint16(len(msgs[i].Buf))
		m[i].buf = uintptr(unsafe.Pointer(&msgs[i].Buf[0]))
		if msgs[i].Flags&I2cFlagRead != 0 {
			m[i].flags |= 0x0001
		}
		if msgs[i].Flags&I2cFlagTenBit != 0 {
			m[i].flags |= 0x0010
		}
	}
	err := ioctl(i2.file.Fd(), i2cRdWr, uintptr(unsafe.Pointer(mi)))
	if err != nil {
		return err
	}
	return nil
}
