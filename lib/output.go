package lib

import (
	"io"
	"sync"
	"time"
)

type Output struct {
	writer       io.Writer
	queue        <-chan []byte
	wg           *sync.WaitGroup
	statsEnabled bool
	resetStats   chan int

	CurrentLines         int64
	TotalLines           int64
	CurrentWriteDuration int64 // micros
	TotalWriteDuration   int64 // micros
}

// Reads from a queue and writes to writer until the queue channel
// is closed. Signals to a WaitGroup when done.
func (output *Output) Write() {
	defer output.wg.Done()

	for {
		select {
		case <-output.resetStats:
			output.CurrentLines = 0
			output.CurrentWriteDuration = 0
		case line := <-output.queue:
			var start time.Time

			if output.statsEnabled {
				start = time.Now()
			}

			output.writer.Write(line)

			if output.statsEnabled {
				output.TotalLines++
				output.CurrentLines++
				delta := time.Now().Sub(start).Nanoseconds() / 1000
				output.CurrentWriteDuration += delta
				output.TotalWriteDuration += delta
			}
		}
	}
}
