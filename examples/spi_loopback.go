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

// Loopback program for SPI.

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aamcrae/gpio"
)

var unit = flag.Int("spi", 0, "SPI Unit")

func main() {
	flag.Parse()
	s, err := io.NewSpi(*unit)
	if err != nil {
		log.Fatalf("Spi unit %d: %v", *unit, err)
	}
	wr_data := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x40, 0x00, 0x00, 0x00, 0x00, 0x95,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xDE, 0xAD, 0xBE, 0xEF, 0xBA, 0xAD,
		0xF0, 0x0D,
	}
	n, err := s.Write(wr_data)
	if err != nil {
		log.Fatalf("write: %v", *unit, err)
	}
	fmt.Printf("Wrote %d bytes", n)
	rb := make([]byte, len(wr_data)*2)
	nb, err := s.Read(rb)
	if err != nil {
		log.Fatalf("read: %v", *unit, err)
	}
	fmt.Printf("Read %d bytes", nb)
	for i := 0; i < nb; i++ {
		fmt.Printf("%02x, ", rb[i])
	}
	fmt.Printf("\n")
	defer s.Close()
}
