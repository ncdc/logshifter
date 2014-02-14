package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type DummyReader struct {
	writeDelay  time.Duration
	msgCount    int
	totalRead   int
	msgTemplate []byte
}

func NewDummyReader(msgCount int, msgLength int, writeDelay time.Duration) *DummyReader {
	msg := []byte(strings.Repeat("0", msgLength-1) + "\n")

	return &DummyReader{writeDelay: writeDelay, msgCount: msgCount, msgTemplate: msg}
}

func (reader *DummyReader) Read(b []byte) (written int, err error) {
	if reader.msgCount == reader.totalRead {
		return 0, fmt.Errorf("EOF: produced %d message(s)", reader.totalRead)
	}

	reader.totalRead++

	copy(b, reader.msgTemplate)

	return len(reader.msgTemplate), nil
}

type DummyWriter struct {
}

func (writer *DummyWriter) Write(b []byte) (written int, err error) {
	return len(b), nil
}

func benchmarkGearLogger(queueSize int, messageCount int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		reader := NewDummyReader(messageCount, 1024, time.Duration(10)*time.Millisecond)
		writer := &DummyWriter{}

		shipper := &GearLogger{queueSize: queueSize, input: reader, writer: writer}

		shipper.Start()
	}
}

// func BenchmarkGearLogger1(b *testing.B) { benchmarkGearLogger(1000, 100, b) }
// func BenchmarkGearLogger2(b *testing.B) { benchmarkGearLogger(1000, 1000, b) }

func BenchmarkGearLogger3(b *testing.B) { benchmarkGearLogger(1000, 10000, b) }

//func BenchmarkGearLogger4(b *testing.B) { benchmarkGearLogger(1000, 100000, b) }
//func BenchmarkGearLogger5(b *testing.B) { benchmarkGearLogger(1000, 1000000, b) }
