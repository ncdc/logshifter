package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	// arg parsing
	var configFile, statsFileName string
	var verbose bool
	var statsInterval time.Duration

	flag.StringVar(&configFile, "config", DefaultConfigFile, "config file location")
	flag.BoolVar(&verbose, "verbose", false, "enables verbose output (e.g. stats reporting)")
	flag.StringVar(&statsFileName, "statsfilename", "", "enabled period stat reporting to the specified file")
	flag.DurationVar(&statsInterval, "statsinterval", (time.Duration(5) * time.Second), "stats reporting interval")
	flag.Parse()

	// load the config
	config, configErr := ParseConfig(configFile)
	if configErr != nil {
		fmt.Printf("Error loading config from %s: %s", configFile, configErr)
		os.Exit(1)
	}

	// override output type from environment if allowed by config
	if config.outputTypeFromEnviron {
		switch os.Getenv("LOGSHIFTER_OUTPUT_TYPE") {
		case "syslog":
			config.outputType = Syslog
		case "file":
			config.outputType = File
		}
	}

	if verbose {
		fmt.Printf("config: %+v\n", config)
	}

	var statsGroup *sync.WaitGroup
	var statsShutdownChan chan int
	var statsChannel chan Stat
	if len(statsFileName) > 0 {
		statsChannel = make(chan Stat)
		statsGroup, statsShutdownChan = readStats(statsChannel, statsInterval, statsFileName)
	}

	writer := createWriter(config)

	shifter := &Shifter{
		queueSize:       config.queueSize,
		inputBufferSize: config.inputBufferSize,
		inputReader:     os.Stdin,
		outputWriter:    writer,
		statsChannel:    statsChannel,
	}

	shifter.Start()

	if statsChannel != nil {
		close(statsChannel)
		statsShutdownChan <- 0
		statsGroup.Wait()
	}
}

func createWriter(config *Config) Writer {
	switch config.outputType {
	case Syslog:
		return &SyslogWriter{config: config}
	case File:
		return nil
	default:
		return nil
	}
}

func readStats(statsChannel chan Stat, interval time.Duration, file string) (wg *sync.WaitGroup, shutdownChan chan int) {
	wg = &sync.WaitGroup{}
	wg.Add(1)

	shutdownChan = make(chan int)

	go func(file string, wg *sync.WaitGroup) {
		defer wg.Done()

		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return
		}

		defer f.Close()

		ticker := time.Tick(interval)
		stats := make(map[string]float64)

		for running := true; running; {
			select {
			case s := <-statsChannel:
				stats[s.name] = stats[s.name] + s.value
			case <-ticker:
				if jsonBytes, err := json.Marshal(stats); err == nil {
					f.Write(jsonBytes)
					f.WriteString("\n")
				}
				stats = make(map[string]float64)
			case <-shutdownChan:
				running = false
			}
		}
	}(file, wg)

	return wg, shutdownChan
}
