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

// Program to read Proximity sensor.

package main

import (
	"flag"
	"log"

	"github.com/aamcrae/gpio"
)

var proxPin = flag.Int("pin", 17, "GPIO pin for sensor")

func main() {
	flag.Parse()
	pin, err := io.Pin(*proxPin)
	if err != nil {
		log.Fatalf("GPIO %d: %v", *proxPin, err)
	}
	p := io.NewProximity(pin)
	last := -1
	for {
		v, err := p.Read()
		if err != nil {
			log.Fatalf("Read: %v", err)
		}
		v = (v - 200) * 100 / 6000
		if v != last {
			log.Printf("Val = %d\n", v)
			last = v
		}
	}
}
