package lib

import (
	"log/syslog"
)

type SyslogWriter struct {
	Config *Config

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
	// TODO: split up using configured buffer size
	return writer.logger.Write(b)
}
