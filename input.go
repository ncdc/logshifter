package main

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"
)

type Input struct {
	bufferSize int
	reader     io.Reader
	queue      chan []byte
	wg         *sync.WaitGroup

	totalLines      int64
	drops           int64
	cumReadDuration int64 // micros
}

// Reads lines from input and writes to queue. If queue is unavailable for
// writing, pops and drops an entry from queue to make room in order to maintain
// a stable consumption rate from input.
//
// Signals to a WaitGroup when there's nothing left to read from input.
func (input *Input) Read() {
	defer input.wg.Done()

	reader := bufio.NewReaderSize(input.reader, input.bufferSize)

	fmt.Println("reader started")

	for {
		line, _, err := reader.ReadLine()

		start := time.Now()

		if err != nil {
			fmt.Println("reader shutting down: ", err)
			break
		}

		input.totalLines++

		select {
		case input.queue <- line:
			// queued
		default:
			// evict the oldest entry to make room
			<-input.queue
			input.drops++
			input.queue <- line
		}

		input.cumReadDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}

	avgReadLatency := float64(input.cumReadDuration) / float64(input.totalLines)

	fmt.Println("total lines read: ", input.totalLines)
	fmt.Println("reader evictions: ", input.drops)
	fmt.Printf("avg read latency (us): %.3v\n", avgReadLatency)
}
