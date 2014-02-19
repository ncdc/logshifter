package lib

import (
	"bufio"
	"io"
	"sync"
	"time"
)

type Input struct {
	bufferSize int
	reader     io.Reader
	queue      chan []byte
	wg         *sync.WaitGroup

	TotalLines      int64
	Drops           int64
	CumReadDuration int64 // micros
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

		if err != nil {
			break
		}

		if len(line) == 0 {
			continue
		}

		cp := make([]byte, len(line))

		copy(cp, line)

		start := time.Now()

		input.TotalLines++

		select {
		case input.queue <- cp:
			// queued
		default:
			// evict the oldest entry to make room
			<-input.queue
			input.Drops++
			input.queue <- cp
		}

		input.CumReadDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}
}
