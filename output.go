package main

import (
	"io"
	"sync"
	"time"
)

type Output struct {
	writer       io.Writer
	queue        <-chan []byte
	statsChannel chan Stat
}

// Reads from a queue and writes to writer until the queue channel
// is closed. Signals to a WaitGroup when done.
func (output *Output) Write() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		output.write()
		wg.Done()
	}(wg)

	return wg
}

func (output *Output) write() {
	for line := range output.queue {
		var start time.Time

		if output.statsChannel != nil {
			start = time.Now()
		}

		output.writer.Write(line)

		if output.statsChannel != nil {
			output.statsChannel <- Stat{name: "output.write", value: 1.0}
			output.statsChannel <- Stat{name: "output.write.duration", value: float64(time.Now().Sub(start).Nanoseconds()) / float64(1000)}
		}
	}
}
