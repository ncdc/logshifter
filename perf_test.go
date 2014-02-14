package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

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

type DummyWriter struct {
}

func (writer *DummyWriter) Write(b []byte) (written int, err error) {
	return len(b), nil
}

func testGearLogger(msgCount, msgLength, lineBufferLength, queueSize int, readerDelay time.Duration, t *testing.T) {
	reader := NewDummyReader(msgCount, msgLength, readerDelay)
	writer := &DummyWriter{}

	logger := &GearLogger{queueSize: queueSize, input: reader, writer: writer, lineBufferLen: lineBufferLength}

	logger.Start()
}

func benchmarkGearLogger(queueSize int, messageCount int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		reader := NewDummyReader(messageCount, 1024, time.Duration(10)*time.Millisecond)
		writer := &DummyWriter{}

		logger := &GearLogger{queueSize: queueSize, input: reader, writer: writer}

		logger.Start()
	}
}

func TestGearLogger1(t *testing.T) {
	testGearLogger(1, 100, 100, 1, 0, t)
}

func TestGearLogger2(t *testing.T) {
	testGearLogger(1, 200, 100, 1, time.Duration(10)*time.Millisecond, t)
}

func TestGearLogger3(t *testing.T) {
	testGearLogger(3, 1, 100, 10, 0, t)
}

// func BenchmarkGearLogger1(b *testing.B) { benchmarkGearLogger(1000, 100, b) }
// func BenchmarkGearLogger2(b *testing.B) { benchmarkGearLogger(1000, 1000, b) }

//func BenchmarkGearLogger3(b *testing.B) { benchmarkGearLogger(1000, 10000, b) }

//func BenchmarkGearLogger4(b *testing.B) { benchmarkGearLogger(1000, 100000, b) }
func BenchmarkGearLogger5(b *testing.B) { benchmarkGearLogger(1000, 1000000, b) }
