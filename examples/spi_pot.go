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

// Digital pot example for SPI.

package main

import (
	"flag"
	"log"
	"time"

	"github.com/aamcrae/gpio"
)

const spiPotStages = 257

var unit = flag.Int("spi", 0, "SPI Unit")

func main() {
	flag.Parse()
	s, err := io.NewSpi(*unit)
	if err != nil {
		log.Fatalf("Spi unit %d: %v", *unit, err)
	}
	defer s.Close()
	for i := 0; i < spiPotStages; i++ {
		Wr(s, i)
	}
	for i := spiPotStages-1; i >= 0; i-- {
		Wr(s, i)
	}
}

func Wr(s *io.Spi, v int) {
	b := []byte{ byte(v >> 8) & 0x3, byte(v & 0xFF)}
	_, err := s.Write(b)
	if err != nil {
		log.Fatalf("write: %v", *unit, err)
	}
	time.Sleep(50 * time.Millisecond)
}
