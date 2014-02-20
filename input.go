package main

import (
	"bufio"
	"io"
	"sync"
	"time"
)

type Input struct {
	bufferSize   int
	reader       io.Reader
	queue        chan []byte
	wg           *sync.WaitGroup
	statsChannel chan Stat
}

// Reads lines from input and writes to queue. If queue is unavailable for
// writing, pops and drops an entry from queue to make room in order to maintain
// a stable consumption rate from input.
//
// Signals to a WaitGroup when there's nothing left to read from input.
func (input *Input) Read() {
	defer input.wg.Done()

	reader := bufio.NewReaderSize(input.reader, input.bufferSize)

	for {
		line, _, err := reader.ReadLine()

		var start time.Time
		if input.statsChannel != nil {
			start = time.Now()
		}

		if err != nil {
			break
		}

		if len(line) == 0 {
			continue
		}

		cp := make([]byte, len(line))

		copy(cp, line)

		if input.statsChannel != nil {
			input.statsChannel <- Stat{name: "input.read", value: 1.0}
		}

		select {
		case input.queue <- cp:
			// queued
		default:
			// evict the oldest entry to make room
			<-input.queue
			if input.statsChannel != nil {
				input.statsChannel <- Stat{name: "input.drop", value: 1.0}
			}
			input.queue <- cp
		}

		if input.statsChannel != nil {
			input.statsChannel <- Stat{name: "input.read.duration", value: float64(time.Now().Sub(start).Nanoseconds()) / float64(1000)}
		}
	}
}
