package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
)

func main() {
	// arg parsing
	var configFile string

	flag.StringVar(&configFile, "config", DefaultConfigFile, "config file location")
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

	fmt.Printf("config: %+v\n", config)

	// create a syslog based input writer
	logger, logErr := syslog.New(syslog.LOG_INFO, "logshifter")

	if logErr != nil {
		fmt.Println("Error opening syslog: %s", logErr)
		os.Exit(1)
	}

	shifter := &Shifter{queueSize: config.queueSize, inputBufferSize: config.inputBufferSize, inputReader: os.Stdin, outputWriter: logger}

	shifter.Start()

	fmt.Println("done.")
}
