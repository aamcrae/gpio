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

	"golang.org/x/sys/unix"
)

var spiDevs = []struct {
	bus, cs byte
}{
	{0, 0},
	{0, 1},
	{1, 0},
	{1, 1},
	{1, 2},
}

// Spi driver modes.
const (
	SPI_MODE_0 = 0
	SPI_MODE_1 = 1
	SPI_MODE_2 = 2
	SPI_MODE_3 = 3

	SPI_MODE_CS_HIGH   = 0x04
	SPI_MODE_LSB_FIRST = 0x08
	// SPI_MODE_3WIRE	   = 0x10	3 wire mode is not supported.
	SPI_MODE_LOOP  = 0x20
	SPI_MODE_NO_CS = 0x40
	SPI_MODE_READY = 0x80
)

var (
	spiXfer       = iocW(0, unsafe.Sizeof(xfer{}))
	spiRdMode     = iocR(1, unsafe.Sizeof(byte(0)))
	spiWrMode     = iocW(1, unsafe.Sizeof(byte(0)))
	spiRdLsbFirst = iocR(2, unsafe.Sizeof(byte(0)))
	spiWrLsbFirst = iocW(2, unsafe.Sizeof(byte(0)))
	spiRdBits     = iocR(3, unsafe.Sizeof(byte(0)))
	spiWrBits     = iocW(3, unsafe.Sizeof(byte(0)))
	spiRdSpeed    = iocR(4, unsafe.Sizeof(uint32(0)))
	spiWrSpeed    = iocW(4, unsafe.Sizeof(uint32(0)))
	spiRdMode32   = iocR(5, unsafe.Sizeof(uint32(0)))
	spiWrMode32   = iocW(5, unsafe.Sizeof(uint32(0)))
)

type Spi struct {
	bus  byte
	cs   byte
	file *os.File
	fd   int
}

type xfer struct {
	txb int64
	rxb int64

	ln    uint32
	speed uint32

	delay    uint16
	bits     byte
	cs       byte
	tx_nbits byte
	rx_nbits byte
	_        uint16
}

// NewSpi creates and initialises a SPI device.
func NewSpi(unit int) (*Spi, error) {
	if unit > len(spiDevs) {
		return nil, os.ErrNotExist
	}
	s := new(Spi)
	s.bus = spiDevs[unit].bus
	s.cs = spiDevs[unit].cs
	var err error
	s.file, err = os.OpenFile(fmt.Sprintf("/dev/spidev%d.%d", s.bus, s.cs), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	s.Speed(100 * 1000)
	s.Bits(8)
	s.Mode(0)
	return s, nil
}

// Xfer builds and sends a message request.
// The receive buffer is returned.
// TODO: The interface will support multiple messages.
func (s *Spi) Xfer(wb []byte) ([]byte, error) {
	x := new(xfer)
	x.txb = int64(uintptr(unsafe.Pointer(&wb[0])))
	rb := make([]byte, len(wb))
	x.rxb = int64(uintptr(unsafe.Pointer(&rb[0])))
	x.ln = uint32(len(wb))
	x.speed = 100 * 1000
	x.bits = 8
	return rb, s.ioctl(spiXfer, uintptr(unsafe.Pointer(x)))
}

// Write writes the message to the SPI device.
func (s *Spi) Write(b []byte) (int, error) {
	return s.file.Write(b)
}

// Read reads a message from the SPI devices
func (s *Spi) Read(b []byte) (int, error) {
	return s.file.Read(b)
}

// Speed sets the speed of the interface.
func (s *Spi) Speed(speed uint32) error {
	return s.ioctl32(spiWrSpeed, &speed)
}

// Bits selects the word size of the transfer (usually 8 or 9 bits)
func (s *Spi) Bits(bits byte) error {
	return s.ioctl8(spiWrBits, &bits)
}

// Mode sets the mode, which is a combination of mode flags.
func (s *Spi) Mode(m uint32) error {
	return s.ioctl32(spiWrMode32, &m)
}

// Close closes the SPI controller
func (s *Spi) Close() {
	s.file.Close()
}

func (s *Spi) ioctl8(req uintptr, b *byte) error {
	return s.ioctl(req, uintptr(unsafe.Pointer(b)))
}

func (s *Spi) ioctl32(req uintptr, i *uint32) error {
	return s.ioctl(req, uintptr(unsafe.Pointer(i)))
}

func (s *Spi) ioctl(req, arg uintptr) error {
	_, _, ep := unix.Syscall(unix.SYS_IOCTL, s.file.Fd(), req, arg)
	if ep != 0 {
		return ep
	}
	return nil
}

func iocR(nr, sz uintptr) uintptr {
	return iocreq(2, nr, sz)
}

func iocW(nr, sz uintptr) uintptr {
	return iocreq(1, nr, sz)
}

func iocreq(dir, nr, sz uintptr) uintptr {
	return (dir << 30) | (sz << 16) | (uintptr('k') << 8) | nr
}
