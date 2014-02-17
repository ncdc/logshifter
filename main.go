package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
)

func main() {
	// arg parsing
	var queueSize, inputBufferSize int

	flag.IntVar(&queueSize, "queuesize", 1000, "max size for the internal line queue")
	flag.IntVar(&inputBufferSize, "inbufsize", 2048, "max length for an input line")
	flag.Parse()

	fmt.Println("queue size ", queueSize)
	fmt.Println("input buffer size ", inputBufferSize)

	// create a syslog based input writer
	logger, logErr := syslog.New(syslog.LOG_INFO, "logshifter")

	if logErr != nil {
		fmt.Println("Error opening syslog: %s", logErr)
		os.Exit(1)
	}

	shifter := &Shifter{queueSize: queueSize, inputBufferSize: inputBufferSize, inputReader: os.Stdin, outputWriter: logger}

	shifter.Start()

	fmt.Println("done.")
}
