package lib

import (
	"fmt"
	"io"
	"sync"
)

type Shifter struct {
	QueueSize       int
	InputBufferSize int
	InputReader     io.Reader
	OutputWriter    Writer
}

type Writer interface {
	io.Writer
	Init() error
}

func (shifter *Shifter) Start() {
	// setup
	queue := make(chan []byte, shifter.QueueSize)
	shifter.OutputWriter.Init()

	readGroup := &sync.WaitGroup{}
	writeGroup := &sync.WaitGroup{}

	readGroup.Add(1)
	writeGroup.Add(1)

	input := &Input{bufferSize: shifter.InputBufferSize, reader: shifter.InputReader, queue: queue, wg: readGroup}
	output := &Output{writer: shifter.OutputWriter, queue: queue, wg: writeGroup}

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
