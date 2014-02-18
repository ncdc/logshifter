package lib

import (
	"io"
	"sync"
	"time"
)

type Output struct {
	writer io.Writer
	queue  <-chan []byte
	wg     *sync.WaitGroup

	TotalLines       int64
	CumWriteDuration int64 // micros
}

// Reads from a queue and writes to writer until the queue channel
// is closed. Signals to a WaitGroup when done.
func (output *Output) Write() {
	defer output.wg.Done()

	for line := range output.queue {
		start := time.Now()

		output.writer.Write(line)

		output.TotalLines++
		output.CumWriteDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}
}
