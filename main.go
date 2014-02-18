package main

import (
	"flag"
	"fmt"
	"github.com/ironcladlou/logshifter/lib"
	"os"
)

func main() {
	// arg parsing
	var configFile string
	var verbose bool

	flag.StringVar(&configFile, "config", lib.DefaultConfigFile, "config file location")
	flag.BoolVar(&verbose, "verbose", false, "enables verbose output (e.g. stats reporting)")
	flag.Parse()

	// load the config
	config, configErr := lib.ParseConfig(configFile)
	if configErr != nil {
		fmt.Printf("Error loading config from %s: %s", configFile, configErr)
		os.Exit(1)
	}

	// override output type from environment if allowed by config
	if config.OutputTypeFromEnviron {
		switch os.Getenv("LOGSHIFTER_OUTPUT_TYPE") {
		case "syslog":
			config.OutputType = lib.Syslog
		case "file":
			config.OutputType = lib.File
		}
	}

	if verbose {
		fmt.Printf("config: %+v\n", config)
	}

	// create a syslog based input writer
	writer := createWriter(config)

	shifter := &lib.Shifter{QueueSize: config.QueueSize, InputBufferSize: config.InputBufferSize, InputReader: os.Stdin, OutputWriter: writer}

	stats := shifter.Start()

	if verbose {
		stats.Print()
	}
}

func createWriter(config *lib.Config) lib.Writer {
	switch config.OutputType {
	case lib.Syslog:
		return &lib.SyslogWriter{Config: config}
	case lib.File:
		return nil
	default:
		return nil
	}
}
