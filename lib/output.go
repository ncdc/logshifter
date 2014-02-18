package lib

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Output struct {
	writer io.Writer
	queue  <-chan []byte
	wg     *sync.WaitGroup

	totalLines       int64
	cumWriteDuration int64 // micros
}

// Reads from a queue and writes to writer until the queue channel
// is closed. Signals to a WaitGroup when done.
func (output *Output) Write() {
	defer output.wg.Done()

	fmt.Println("writer started")

	for line := range output.queue {
		start := time.Now()

		output.writer.Write(line)

		output.totalLines++
		output.cumWriteDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}

	fmt.Println("writer shutting down")

	avgWriteDuration := float64(output.cumWriteDuration) / float64(output.totalLines)

	fmt.Println("total lines written: ", output.totalLines)
	fmt.Printf("avg write duration (us): %.3v\n", avgWriteDuration)
}
