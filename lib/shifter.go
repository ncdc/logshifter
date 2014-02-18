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

type Stats struct {
	InputLinesTotal        int64
	InputDrops             int64
	InputCumReadDuration   int64
	InputAvgReadLatency    float64
	OutputLinesTotal       int64
	OutputCumWriteDuration int64
	OutputAvgWriteDuration float64
}

type Writer interface {
	io.Writer
	Init() error
}

func (shifter *Shifter) Start() Stats {
	// setup
	queue := make(chan []byte, shifter.QueueSize)
	shifter.OutputWriter.Init()

	readGroup := &sync.WaitGroup{}
	writeGroup := &sync.WaitGroup{}

	readGroup.Add(1)
	writeGroup.Add(1)

	input := &Input{bufferSize: shifter.InputBufferSize, reader: shifter.InputReader, queue: queue, wg: readGroup}
	output := &Output{writer: shifter.OutputWriter, queue: queue, wg: writeGroup}

	// start writing before reading: there's still a race here, not worth bothering with yet
	go output.Write()
	go input.Read()

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()

	stats := Stats{}

	stats.InputLinesTotal = input.TotalLines
	stats.InputDrops = input.Drops
	stats.InputCumReadDuration = input.CumReadDuration
	stats.InputAvgReadLatency = float64(stats.InputCumReadDuration) / float64(stats.InputLinesTotal)

	stats.OutputLinesTotal = output.TotalLines
	stats.OutputCumWriteDuration = output.CumWriteDuration
	stats.OutputAvgWriteDuration = float64(output.CumWriteDuration) / float64(output.TotalLines)

	return stats
}

func (stats *Stats) Print() {
	fmt.Printf("total lines read: %d\n", stats.InputLinesTotal)
	fmt.Printf("reader evictions: %d\n", stats.InputDrops)
	fmt.Printf("avg read latency (us): %.3v\n", stats.InputAvgReadLatency)
	fmt.Printf("total lines written: %d\n", stats.OutputLinesTotal)
	fmt.Printf("avg write duration (us): %.3v\n", stats.OutputAvgWriteDuration)
}
