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

package sensor

import (
	"time"

	"github.com/aamcrae/gpio"
)

// Proximity represents a driver for the QRE1113 proximity sensor, configured
// as a digital circuit.
// This operates by setting the pin as an output, charging a capacitor,
// and then switching the pin to an input and waiting for the capacitor to
// discharge through the sensor. The faster the discharge time, the
// stronger the reflected signal.
type Proximity struct {
	pin      *io.Gpio // Pin for reading and controlling reader.
	Min, Max int      // For range checks
}

// NewProximity creates and initialises a Proximity struct.
func NewProximity(pin *io.Gpio) *Proximity {
	p := &Proximity{pin, 100, 5000}
	p.pin.Direction(io.IN)
	p.pin.Edge(io.FALLING)
	return p
}

// Read triggers the sensor by setting the output pin high to charge the capacitor,
// then turning it off, and detecting how long the capacitor takes to drain.
// The duration is returned as the number of microseconds.
func (p *Proximity) Read() (int, error) {
	for retries := 0; retries < 5; retries++ {
		p.pin.Direction(io.OUT)
		p.pin.Set(1)
		time.Sleep(time.Microsecond * 100)
		now := time.Now()
		p.pin.Direction(io.IN)
		for {
			v, err := p.pin.GetTimeout(time.Millisecond * 20)
			if err != nil {
				return 0, err
			}
			if v == 0 {
				diff := int(time.Now().Sub(now).Microseconds())
				// If out of range, try again.
				if diff < p.Min || diff > p.Max {
					break
				}
				return diff, nil
			}
		}
	}
	return 0, io.ErrRetriesExceeded
}
