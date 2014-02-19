package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ironcladlou/logshifter/lib"
)

func main() {
	// arg parsing
	var configFile, statsFileName string
	var verbose bool
	var statsInterval time.Duration

	flag.StringVar(&configFile, "config", lib.DefaultConfigFile, "config file location")
	flag.BoolVar(&verbose, "verbose", false, "enables verbose output (e.g. stats reporting)")
	flag.StringVar(&statsFileName, "statsfilename", "", "enabled period stat reporting to the specified file")
	flag.DurationVar(&statsInterval, "statsinterval", (time.Duration(5) * time.Second), "stats reporting interval")
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

	var statsChannel chan lib.Stats
	var statsWaitGroup *sync.WaitGroup
	if len(statsFileName) > 0 {
		statsChannel, statsWaitGroup = createStatsChannel(statsFileName)
	}

	// create a syslog based input writer
	writer := createWriter(config)

	shifter := &lib.Shifter{QueueSize: config.QueueSize, InputBufferSize: config.InputBufferSize, InputReader: os.Stdin, OutputWriter: writer, StatsChannel: statsChannel, StatsInterval: statsInterval}

	stats := shifter.Start()

	if statsChannel != nil {
		close(statsChannel)
		statsWaitGroup.Wait()
	}

	if verbose && statsChannel != nil {
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

func createStatsChannel(file string) (chan lib.Stats, *sync.WaitGroup) {
	c := make(chan lib.Stats)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(file string, wg *sync.WaitGroup) {
		defer wg.Done()

		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return
		}

		for stat := range c {
			if jsonBytes, err := json.Marshal(stat); err == nil {
				f.Write(jsonBytes)
				f.WriteString("\n")
			}
		}

		f.Close()
	}(file, wg)

	return c, wg
}
