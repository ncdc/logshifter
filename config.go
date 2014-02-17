package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// output types
	Syslog = "syslog"
	File   = "file"

	DefaultConfigFile = "/etc/openshift/logshifter.conf"
)

type Config struct {
	queueSize             int
	inputBufferSize       int
	outputType            string
	syslogBufferSize      int
	fileBufferSize        int
	outputTypeFromEnviron bool
}

func ParseConfig(file string) (*Config, error) {
	fmt.Println("loading config from ", file)

	config := &Config{}

	f, err := os.Open(file)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil || len(line) == 0 {
			break
		}

		c := strings.SplitN(line, "=", 2)

		if len(c) != 2 {
			break
		}

		k := strings.Trim(c[0], "\n ")
		v := strings.Trim(c[1], "\n ")

		switch k {
		case "queuesize":
			config.queueSize, _ = strconv.Atoi(v)
		case "inputbuffersize":
			config.inputBufferSize, _ = strconv.Atoi(v)
		case "outputtype":
			switch v {
			case "syslog":
				config.outputType = Syslog
			case "file":
				config.outputType = File
			}
		case "syslogbuffersize":
			config.syslogBufferSize, _ = strconv.Atoi(v)
		case "filebuffersize":
			config.fileBufferSize, _ = strconv.Atoi(v)
		case "outputtypefromenviron":
			config.outputTypeFromEnviron, _ = strconv.ParseBool(v)
		}
	}

	return config, nil
}
