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

	input            *Input
	output           *Output
	resetInputStats  chan int
	resetOutputStats chan int
}

type Stats struct {
	InputLinesCurrent             int64
	InputLinesTotal               int64
	InputDropsCurrent             int64
	InputDropsTotal               int64
	InputReadDurationCurrent      int64
	InputReadDurationTotal        int64
	InputAvgReadLatencyCurrent    float64
	InputAvgReadLatencyTotal      float64
	OutputLinesTotal              int64
	OutputLinesCurrent            int64
	OutputWriteDurationCurrent    int64
	OutputWriteDurationTotal      int64
	OutputAvgWriteDurationCurrent float64
	OutputAvgWriteDurationTotal   float64
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

	shifter.resetInputStats = make(chan int)
	shifter.resetOutputStats = make(chan int)

	input := &Input{
		bufferSize:   shifter.InputBufferSize,
		reader:       shifter.InputReader,
		queue:        queue,
		wg:           readGroup,
		statsEnabled: shifter.StatsEnabled,
		resetStats:   shifter.resetInputStats,
	}

	output := &Output{
		writer:       shifter.OutputWriter,
		queue:        queue,
		wg:           writeGroup,
		statsEnabled: shifter.StatsEnabled,
		resetStats:   shifter.resetOutputStats,
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
			shifter.resetInputStats <- 1
			shifter.resetOutputStats <- 1
		}
	}
}

func (shifter *Shifter) buildStats() Stats {
	stats := Stats{}

	stats.InputLinesTotal = shifter.input.TotalLines
	stats.InputLinesCurrent = shifter.input.CurrentLines
	stats.InputDropsCurrent = shifter.input.CurrentDrops
	stats.InputDropsTotal = shifter.input.TotalDrops
	stats.InputReadDurationCurrent = shifter.input.CurrentReadDuration
	stats.InputReadDurationTotal = shifter.input.TotalReadDuration
	if stats.InputLinesCurrent > 0 {
		stats.InputAvgReadLatencyCurrent = float64(stats.InputReadDurationCurrent) / float64(stats.InputLinesCurrent)
	} else {
		stats.InputAvgReadLatencyCurrent = 0
	}

	if stats.InputLinesTotal > 0 {
		stats.InputAvgReadLatencyTotal = float64(stats.InputReadDurationTotal) / float64(stats.InputLinesTotal)
	} else {
		stats.InputAvgReadLatencyTotal = 0
	}

	stats.OutputLinesCurrent = shifter.output.CurrentLines
	stats.OutputLinesTotal = shifter.output.TotalLines
	stats.OutputWriteDurationCurrent = shifter.output.CurrentWriteDuration
	stats.OutputWriteDurationTotal = shifter.output.TotalWriteDuration
	if stats.OutputLinesCurrent > 0 {
		stats.OutputAvgWriteDurationCurrent = float64(stats.OutputWriteDurationCurrent) / float64(stats.OutputLinesCurrent)
	} else {
		stats.OutputAvgWriteDurationCurrent = 0
	}
	if stats.OutputLinesTotal > 0 {
		stats.OutputAvgWriteDurationTotal = float64(stats.OutputWriteDurationTotal) / float64(stats.OutputLinesTotal)
	} else {
		stats.OutputAvgWriteDurationTotal = 0
	}

	return stats
}
