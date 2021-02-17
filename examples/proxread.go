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
	"fmt"
	"log"

	"github.com/aamcrae/gpio"
)

var proxPin = flag.Int("pin", 21, "GPIO pin for sensor")
var proxMin = flag.Int("min", 200, "Minimum value allowed")
var proxMax = flag.Int("max", 2500, "Maximum value allowed")

func main() {
	flag.Parse()
	pin, err := io.Pin(*proxPin)
	if err != nil {
		log.Fatalf("GPIO %d: %v", *proxPin, err)
	}
	p := io.NewProximity(pin)
	p.Min = *proxMin
	p.Max = *proxMax
	last := -1
	lastSeg := -1
	max := -1
	min := 1_000_000
	for {
		v, err := p.Read()
		if err != nil {
			log.Fatalf("Read: %v", err)
		}
		if v != last {
			if v > max {
				max = v
			}
			if v < min {
				min = v
			}
			if max == min {
				continue
			}
			s := (v - min) * 100 / (max - min)
			if lastSeg == s/10 {
				continue
			}
			fmt.Printf("Value: %5d, min: %5d, max %5d", v, min, max)
			fmt.Printf(", scaled %3d |", s)
			var i int
			for ; i < s/10; i++ {
				fmt.Printf("-")
			}
			for ; i < 10; i++ {
				fmt.Printf(" ")
			}
			fmt.Printf("|\n")
			last = v
			lastSeg = s/10
		}
	}
}
