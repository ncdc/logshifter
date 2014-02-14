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
	fmt.Println("b len=", len(b))
	if reader.msgCount == reader.totalRead {
		return 0, fmt.Errorf("EOF: reader produced %d messages", reader.totalRead)
	}

	reader.totalRead++

	copy(reader.msgTemplate, b)

	return len(reader.msgTemplate), nil
}

type DummyWriter struct {
}

func (writer *DummyWriter) Write(b []byte) (written int, err error) {
	return len(b), nil
}

func TestStuff(t *testing.T) {
	var queueSize int = 1000
	reader := NewDummyReader(10, 50, time.Duration(10)*time.Millisecond)
	writer := &DummyWriter{}

	shipper := &GearLogger{queueSize: queueSize, input: reader, writer: writer}

	shipper.Start()
}
