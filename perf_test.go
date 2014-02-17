package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Simulates an upstream reader of a producer such as stdin
type DummyReader struct {
	buffer      *bytes.Buffer
	readerDelay time.Duration
}

func NewDummyReader(msgCount int, msgLength int, readerDelay time.Duration) *DummyReader {
	var data string

	for i := 0; i < msgCount; i++ {
		data += strings.Repeat("0", msgLength-1) + "\n"
	}

	buffer := bytes.NewBufferString(data)

	fmt.Printf("created %d byte test input\n", buffer.Len())

	return &DummyReader{buffer: buffer, readerDelay: readerDelay}
}

func (reader *DummyReader) Read(b []byte) (written int, err error) {
	if reader.readerDelay > 0 {
		time.Sleep(reader.readerDelay)
	}

	written, err = reader.buffer.Read(b)

	return
}

// Simulates a downstream log writer such as syslog
type DummyWriter struct {
	writerDelay time.Duration
}

func (writer *DummyWriter) Write(b []byte) (written int, err error) {
	if writer.writerDelay > 0 {
		time.Sleep(writer.writerDelay)
	}

	return len(b), nil
}

func testShifter(msgCount, msgLength, inputBufferSize, queueSize int, readerDelay time.Duration, t *testing.T) {
	reader := NewDummyReader(msgCount, msgLength, readerDelay)
	writer := &DummyWriter{}

	shifter := &Shifter{queueSize: queueSize, inputBufferSize: inputBufferSize, inputReader: reader, outputWriter: writer}

	shifter.Start()
}

func benchmarkShifter(queueSize, msgLength, messageCount int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		reader := NewDummyReader(messageCount, msgLength, 0)
		writer := &DummyWriter{}

		shifter := &Shifter{queueSize: queueSize, inputBufferSize: msgLength, inputReader: reader, outputWriter: writer}

		shifter.Start()
	}
}

func TestShifter1(t *testing.T) {
	testShifter(1, 100, 100, 1, 0, t)
}

func TestShifter2(t *testing.T) {
	testShifter(1, 200, 100, 1, time.Duration(10)*time.Millisecond, t)
}

func TestShifter3(t *testing.T) {
	testShifter(3, 1, 100, 10, 0, t)
}

func BenchmarkShifter(b *testing.B) { benchmarkShifter(1000, 100, 10000, b) }
