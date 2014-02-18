package main

import (
	"flag"
	"fmt"
	"github.com/ironcladlou/logshifter/lib"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	// arg parsing
	var configFile, statsFile string
	var verbose bool
	var statsInterval time.Duration

	flag.StringVar(&configFile, "config", lib.DefaultConfigFile, "config file location")
	flag.BoolVar(&verbose, "verbose", false, "enables verbose output (e.g. stats reporting)")
	flag.StringVar(&statsFile, "stats", "", "enabled period stat reporting to the specified file")
	flag.DurationVar(&statsInterval, "statsint", (time.Duration(5) * time.Second), "stats reporting interval")
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
	if len(statsFile) > 0 {
		c, err := createStatsChannel(statsFile)
		if err != nil {
			fmt.Printf("Error opening stats file %s: %s", statsFile, err)
			os.Exit(1)
		}

		statsChannel = c
	}

	// create a syslog based input writer
	writer := createWriter(config)

	shifter := &lib.Shifter{QueueSize: config.QueueSize, InputBufferSize: config.InputBufferSize, InputReader: os.Stdin, OutputWriter: writer, StatsChannel: statsChannel, StatsInterval: statsInterval}

	stats := shifter.Start()

	if statsChannel != nil {
		close(statsChannel)
	}

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

func createStatsChannel(file string) (chan lib.Stats, error) {
	// verify we can write to the target file
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	f.Close()

	c := make(chan lib.Stats)

	go func() {
		for stat := range c {
			ioutil.WriteFile(file, []byte(stat.ToString()), os.ModePerm)
		}
	}()

	return c, nil
}
