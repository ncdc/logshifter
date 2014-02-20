package main

import (
	"fmt"
	"testing"
)

func TestSimpleDispatch(t *testing.T) {
	var msgCount int64 = 1000

	statsChan := make(chan Stat)

	reader := NewDummyReader(msgCount, 100, 0)
	writer := &DummyWriter{}
	shifter := &Shifter{
		queueSize:       1000,
		inputBufferSize: 100,
		inputReader:     reader,
		outputWriter:    writer,
		statsChannel:    statsChan,
	}

	stats := make(map[string]float64)
	go func() {
		for s := range statsChan {
			stats[s.name] = stats[s.name] + s.value
		}
	}()

	shifter.Start()

	close(statsChan)

	fmt.Printf("stats: %v\n", stats)
}
