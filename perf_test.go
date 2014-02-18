package main

import (
	"github.com/ironcladlou/logshifter/lib"
	"testing"
)

func benchmarkShifter(queueSize int, msgLength int, messageCount int64, b *testing.B) {
	for n := 0; n < b.N; n++ {
		reader := NewDummyReader(messageCount, msgLength, 0)
		writer := &DummyWriter{}

		shifter := &lib.Shifter{QueueSize: queueSize, InputBufferSize: msgLength, InputReader: reader, OutputWriter: writer}

		shifter.Start()
	}
}

func BenchmarkShifter(b *testing.B) { benchmarkShifter(1000, 100, 10000, b) }
