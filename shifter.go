package main

import (
	"io"
	"sync"
)

type Shifter struct {
	queueSize       int
	inputBufferSize int
	inputReader     io.Reader
	outputWriter    Writer
	statsChannel    chan Stat

	input  *Input
	output *Output
}

type Writer interface {
	io.Writer
	Init() error
}

type Stat struct {
	name  string
	value float64
}

func (shifter *Shifter) Start() {
	// setup
	queue := make(chan []byte, shifter.queueSize)
	shifter.outputWriter.Init()

	readGroup := &sync.WaitGroup{}
	writeGroup := &sync.WaitGroup{}

	readGroup.Add(1)
	writeGroup.Add(1)

	input := &Input{
		bufferSize:   shifter.inputBufferSize,
		reader:       shifter.inputReader,
		queue:        queue,
		wg:           readGroup,
		statsChannel: shifter.statsChannel,
	}

	output := &Output{
		writer:       shifter.outputWriter,
		queue:        queue,
		wg:           writeGroup,
		statsChannel: shifter.statsChannel,
	}

	shifter.input = input
	shifter.output = output

	// start writing before reading: there's still a race here, not worth bothering with yet
	go shifter.output.Write()
	go shifter.input.Read()

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()
}
