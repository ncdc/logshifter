package main

import (
	"github.com/ironcladlou/logshifter/lib"
	"testing"
)

func TestSimpleDispatch(t *testing.T) {
	var msgCount int64 = 1000

	reader := NewDummyReader(msgCount, 100, 0)
	writer := &DummyWriter{}
	shifter := &lib.Shifter{QueueSize: 1000, InputBufferSize: 100, InputReader: reader, OutputWriter: writer}

	stats := shifter.Start()

	if stats.InputLinesTotal != msgCount {
		t.Fatalf("expected %d input lines, got %d", msgCount, stats.InputLinesTotal)
	}

	if stats.OutputLinesTotal != msgCount {
		t.Fatalf("expected %d output lines, got %d", msgCount, stats.OutputLinesTotal)
	}

	stats.Print()
}

// Test to ensure that lines exceeding the input buffer size are split
// into separate lines in the queue.
func TestInputBufferOverflow(t *testing.T) {
	var msgCount int64 = 1000

	reader := NewDummyReader(msgCount, 200, 0)
	writer := &DummyWriter{}
	shifter := &lib.Shifter{QueueSize: 2000, InputBufferSize: 100, InputReader: reader, OutputWriter: writer}

	stats := shifter.Start()

	if stats.InputLinesTotal != 2000 {
		t.Fatalf("expected %d input lines, got %d", msgCount, stats.InputLinesTotal)
	}

	if stats.OutputLinesTotal != 2000 {
		t.Fatalf("expected %d output lines, got %d", msgCount, stats.OutputLinesTotal)
	}

	stats.Print()
}
