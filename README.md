# gpio
Go GPIO library.

The base library contains interface code to drivers and hardware.

The ```action``` directory contains types that use the lower layer interface
code to build more complex actions e.g a stepper motor driver that uses individual
GPIO elements to control the different inputs to the motor.

The ```sensor``` directory contains types that provide a higher level
interface to sensors that are accessed via the lower layer interface types.

The ```examples``` directory contains sample programs that demonstrate
the use of the library.
