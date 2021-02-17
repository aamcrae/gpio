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

var (
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
	bus   byte
	cs    byte
	file  *os.File
	fd    int
	speed int // In Hz
	bits  int
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
	return s, nil
}

func (s *Spi) Write(b []byte) (int, error) {
	return s.file.Write(b)
}

func (s *Spi) Read(b []byte) (int, error) {
	return s.file.Read(b)
}

func (s *Spi) Speed(speed uint32) error {
	return s.ioctl32(spiWrSpeed, &speed)
}

func (s *Spi) Bits(bits byte) error {
	return s.ioctl8(spiWrBits, &bits)
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
