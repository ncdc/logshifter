package main

import (
	"io"
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

	shifter.input = &Input{
		bufferSize:   shifter.inputBufferSize,
		reader:       shifter.inputReader,
		queue:        queue,
		statsChannel: shifter.statsChannel,
	}

	shifter.output = &Output{
		writer:       shifter.outputWriter,
		queue:        queue,
		statsChannel: shifter.statsChannel,
	}

	// start writing before reading: there's still a race here, not worth bothering with yet
	writeGroup := shifter.output.Write()
	readGroup := shifter.input.Read()

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()
}
