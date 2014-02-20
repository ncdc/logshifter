package main

import (
	"bytes"
	"log/syslog"
)

type SyslogWriter struct {
	config *Config

	logger *syslog.Writer
}

func (writer *SyslogWriter) Init() error {
	logger, err := syslog.New(syslog.LOG_INFO, "logshifter")

	if err != nil {
		return err
	}

	writer.logger = logger

	return nil
}

func (writer *SyslogWriter) Write(b []byte) (n int, err error) {
	if len(b) > writer.config.syslogBufferSize {
		// Break up messages that exceed the downstream buffer length,
		// using a bytes.Buffer since it's easy. This may result in an
		// undesirable amount of allocations, but the assumption is that
		// bursts of too-long messages are rare.
		buf := bytes.NewBuffer(b)
		for buf.Len() > 0 {
			writer.logger.Write(buf.Next(writer.config.syslogBufferSize))
		}

		return len(b), nil
	} else {
		return writer.logger.Write(b)
	}
}
