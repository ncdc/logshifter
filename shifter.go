package main

import (
	"fmt"
	"io"
	"sync"
)

type Shifter struct {
	queueSize       int
	inputBufferSize int
	inputReader     io.Reader
	outputWriter    io.Writer
}

func (shifter *Shifter) Start() {
	// setup
	queue := make(chan []byte, shifter.queueSize)

	readGroup := &sync.WaitGroup{}
	writeGroup := &sync.WaitGroup{}

	readGroup.Add(1)
	writeGroup.Add(1)

	input := &Input{bufferSize: shifter.inputBufferSize, reader: shifter.inputReader, queue: queue, wg: readGroup}
	output := &Output{writer: shifter.outputWriter, queue: queue, wg: writeGroup}

	// start writing before reading: there's still a race here, but not a problem
	// for this POC.
	go output.Write()
	go input.Read()

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()

	fmt.Println("logshifter finished")
}
