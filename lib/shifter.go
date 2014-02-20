package lib

import (
	"io"
	"sync"
	"time"
)

type Shifter struct {
	QueueSize       int
	InputBufferSize int
	InputReader     io.Reader
	OutputWriter    Writer
	StatsEnabled    bool
	StatsChannel    chan Stats
	StatsInterval   time.Duration

	input  *Input
	output *Output
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

	input := &Input{
		bufferSize:   shifter.InputBufferSize,
		reader:       shifter.InputReader,
		queue:        queue,
		wg:           readGroup,
		statsEnabled: shifter.StatsEnabled,
	}

	output := &Output{
		writer:       shifter.OutputWriter,
		queue:        queue,
		wg:           writeGroup,
		statsEnabled: shifter.StatsEnabled,
	}

	shifter.input = input
	shifter.output = output

	// start writing before reading: there's still a race here, not worth bothering with yet
	go shifter.output.Write()
	go shifter.input.Read()

	if shifter.StatsEnabled && shifter.StatsChannel != nil {
		go shifter.reportStats()
	}

	// wait for the the reader to complete
	readGroup.Wait()

	// shut down the writer by closing the queue
	close(queue)
	writeGroup.Wait()

	return shifter.buildStats()
}

func (shifter *Shifter) reportStats() {
	ticker := time.Tick(shifter.StatsInterval)

	for {
		select {
		case <-ticker:
			shifter.StatsChannel <- shifter.buildStats()
		}
	}
}

func (shifter *Shifter) buildStats() Stats {
	stats := Stats{}

	stats.InputLinesTotal = shifter.input.TotalLines
	stats.InputDrops = shifter.input.Drops
	stats.InputCumReadDuration = shifter.input.CumReadDuration
	stats.InputAvgReadLatency = float64(stats.InputCumReadDuration) / float64(stats.InputLinesTotal)

	stats.OutputLinesTotal = shifter.output.TotalLines
	stats.OutputCumWriteDuration = shifter.output.CumWriteDuration
	stats.OutputAvgWriteDuration = float64(stats.OutputCumWriteDuration) / float64(stats.OutputLinesTotal)

	return stats
}
