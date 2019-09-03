package util

import (
	"time"
)

// Used for simple benchmarking.
type StopWatch struct {
	startTime int64
	stopTime  int64
	running   bool
}

// Start the timer.
func (this *StopWatch) Start() {
	this.startTime = time.Now().UnixNano()
	this.running = true
}

// Stop the timer.
func (this *StopWatch) Stop() {
	this.stopTime = time.Now().UnixNano()
	this.running = false
}

// Get the timer's elapsed time.
func (this *StopWatch) GetElapsed() time.Duration {
	if this.running {
		return time.Duration(time.Now().UnixNano() - this.startTime)
	} else {
		return time.Duration(this.stopTime - this.startTime)
	}
}

// Reset the timer.
func (this *StopWatch) Reset() {
	this.startTime = 0
	this.stopTime = 0
	this.running = false
}
