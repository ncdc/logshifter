package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"sync"
	"time"
)

type GearLogger struct {
	queueSize     int
	lineBufferLen int
	writer        io.Writer
	input         io.Reader
}

func (s *GearLogger) Start() {
	// setup
	queue := make(chan []byte, s.queueSize)

	readGroup := &sync.WaitGroup{}
	writeGroup := &sync.WaitGroup{}

	readGroup.Add(1)
	writeGroup.Add(1)

	// start writing before reading: there's still a race here, but not a problem
	// for this POC.
	go s.Write(queue, s.writer, writeGroup)
	go s.Read(s.input, queue, readGroup)

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()

	fmt.Println("gearlogger shutting down")
}

// Reads lines from input and writes to queue. If queue is unavailable for
// writing, pops and drops an entry from queue to make room in order to maintain
// a stable consumption rate from input.
//
// Signals to a WaitGroup when there's nothing left to read from input.
func (s *GearLogger) Read(input io.Reader, queue chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReaderSize(input, s.lineBufferLen)

	fmt.Println("reader started")

	var totalLines int64 = 0
	var drops int64 = 0
	var cumuReadDuration int64 = 0 // micros

	for {
		line, _, err := reader.ReadLine()

		start := time.Now()

		if err != nil {
			fmt.Println("reader shutting down: ", err)
			break
		}

		totalLines++

		select {
		case queue <- line:
			// queued
		default:
			// evict the oldest entry to make room
			<-queue
			drops++
			queue <- line
		}

		cumuReadDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}

	avgReadLatency := float64(cumuReadDuration) / float64(totalLines)

	fmt.Println("total lines read: ", totalLines)
	fmt.Println("reader evictions: ", drops)
	fmt.Printf("avg read latency (us): %.3v\n", avgReadLatency)
}

// Reads from a queue and writes to writer until the queue channel
// is closed. Signals to a WaitGroup when done.
func (s *GearLogger) Write(queue chan []byte, writer io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("writer started")

	var totalWrites int64 = 0
	var cumuWriteDuration int64 = 0

	for line := range queue {
		start := time.Now()

		writer.Write(line)

		totalWrites++
		cumuWriteDuration += time.Now().Sub(start).Nanoseconds() / 1000
	}

	fmt.Println("writer shutting down")

	avgWriteDuration := float64(cumuWriteDuration) / float64(totalWrites)

	fmt.Println("total lines written: ", totalWrites)
	fmt.Printf("avg write duration (us): %.3v\n", avgWriteDuration)
}

func main() {
	// syslog setup
	logger, logErr := syslog.New(syslog.LOG_INFO, "openshift-gear-logger")

	if logErr != nil {
		fmt.Println("Error opening syslog: %s", logErr)
		os.Exit(1)
	}

	// arg parsing
	var queueSize int

	flag.IntVar(&queueSize, "queuesize", 1000, "max size for the internal line queue")
	flag.Parse()

	fmt.Println("using queue size ", queueSize)

	gearLogger := &GearLogger{queueSize: queueSize, input: os.Stdin, writer: logger, lineBufferLen: 2048}

	gearLogger.Start()

	fmt.Println("done.")
}
