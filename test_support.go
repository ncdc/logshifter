package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// Simulates an upstream reader of a producer such as stdin
type DummyReader struct {
	buffer      *bytes.Buffer
	readerDelay time.Duration
}

func NewDummyReader(msgCount int64, msgLength int, readerDelay time.Duration) *DummyReader {
	var data string

	var i int64 = 0
	for ; i < msgCount; i++ {
		data += strings.Repeat("0", msgLength-1) + "\n"
	}

	buffer := bytes.NewBufferString(data)

	fmt.Printf("created %d byte test input (%d lines @ %d bytes each)\n", buffer.Len(), msgCount, msgLength)

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

func (writer *DummyWriter) Init() error { return nil }

func (writer *DummyWriter) Write(b []byte) (written int, err error) {
	if writer.writerDelay > 0 {
		time.Sleep(writer.writerDelay)
	}

	return len(b), nil
}
