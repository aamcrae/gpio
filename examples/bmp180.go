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

// BMP180 sensor, accessed via i2c

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aamcrae/gpio"
)

const spiPotStages = 257

var bus = flag.Int("bus", 1, "I2C bus number")
var addr = flag.Int("addr", 0x77, "I2C device address")

type cal_params struct {
	AC1 int16
	AC2 int16
	AC3 int16
	AC4 uint16
	AC5 uint16
	AC6 uint16
	B1  int16
	B2  int16
	MB  int16
	MC  int16
	MD  int16
}

func main() {
	flag.Parse()
	i2, err := io.NewI2C(*bus)
	if err != nil {
		log.Fatalf("I2C bus %d: %v", *bus, err)
	}
	defer i2.Close()
	i2.Addr(0x77)
	b := make([]byte, 22)
	err = i2.Read(0xAA, b)
	if err != nil {
		log.Fatal("I2C bus %d, message: %v", *bus, err)
	}
	// Data is big endian.
	buf := bytes.NewReader(b)
	var p cal_params
	binary.Read(buf, binary.BigEndian, &p.AC1)
	binary.Read(buf, binary.BigEndian, &p.AC2)
	binary.Read(buf, binary.BigEndian, &p.AC3)
	binary.Read(buf, binary.BigEndian, &p.AC4)
	binary.Read(buf, binary.BigEndian, &p.AC5)
	binary.Read(buf, binary.BigEndian, &p.AC6)
	binary.Read(buf, binary.BigEndian, &p.B1)
	binary.Read(buf, binary.BigEndian, &p.B2)
	binary.Read(buf, binary.BigEndian, &p.MB)
	binary.Read(buf, binary.BigEndian, &p.MC)
	binary.Read(buf, binary.BigEndian, &p.MD)
	fmt.Printf("AC6 = %d, AC5 = %d, MC = %d, MD = %d\n", p.AC6, p.AC5, p.MC, p.MD)
	i2.WriteReg(0xF4, 0x2E)
	time.Sleep(5 * time.Millisecond)
	b = make([]byte, 2)
	i2.Read(0xF6, b)
	UT := binary.BigEndian.Uint16(b)
	oss := byte(0)
	i2.WriteReg(0xF4, 0x34+(oss<<6))
	time.Sleep(5 * time.Millisecond)
	b = make([]byte, 3)
	i2.Read(0xF6, b)
	UP := ((uint(b[0]) << 16) | (uint(b[1]) << 8) | uint(b[2])) >> (8 - oss)
	x1 := int((int64(UT) - int64(p.AC6)) * int64(p.AC5) / (1 << 15))
	x2 := int(p.MC) * (1 << 11) / (int(x1) + int(p.MD))
	b5 := x1 + x2
	t := (b5 + 8) / (1 << 4)
	fmt.Printf("t = %d, UP = %d, UT = %d, b5 = %d, x1 = %d, x2 = %d\n", t, UP, UT, b5, x1, x2)
	fmt.Printf("Temperature = %.1f degrees\n", float64(t)/10)
}
