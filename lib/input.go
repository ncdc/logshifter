package lib

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
	statsEnabled bool
	resetStats   chan int

	CurrentLines        int64
	TotalLines          int64
	CurrentDrops        int64
	TotalDrops          int64
	CurrentReadDuration int64 // micros
	TotalReadDuration   int64 // micros
}

// Reads lines from input and writes to queue. If queue is unavailable for
// writing, pops and drops an entry from queue to make room in order to maintain
// a stable consumption rate from input.
//
// Signals to a WaitGroup when there's nothing left to read from input.
func (input *Input) Read() {
	defer input.wg.Done()

	go func() {
		for {
			select {
			case <-input.resetStats:
				input.CurrentLines = 0
				input.CurrentDrops = 0
				input.CurrentReadDuration = 0
			}
		}
	}()

	reader := bufio.NewReaderSize(input.reader, input.bufferSize)

	for {
		line, _, err := reader.ReadLine()
		var start time.Time
		if input.statsEnabled {
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

		if input.statsEnabled {
			input.CurrentLines++
			input.TotalLines++
		}

		select {
		case input.queue <- cp:
			// queued
		default:
			// evict the oldest entry to make room
			<-input.queue
			if input.statsEnabled {
				input.CurrentDrops++
				input.TotalDrops++
			}
			input.queue <- cp
		}

		if input.statsEnabled {
			delta := time.Now().Sub(start).Nanoseconds() / 1000
			input.CurrentReadDuration += delta
			input.TotalReadDuration += delta
		}
	}
}
