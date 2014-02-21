package main

import (
	"bytes"
	"os"
	"os/user"
	"path"
	"strings"
)

const (
	NL = "\n"
)

type FileWriter struct {
	config *Config
	file   *os.File
	tag    string
}

func (writer *FileWriter) Init() error {
	basedir := writer.config.fileWriterDir

	usr, _ := user.Current()
	dir := usr.HomeDir
	if basedir[:2] == "~/" {
		basedir = strings.Replace(basedir, "~/", (dir + "/"), 1)
	}

	filename := path.Join(basedir, writer.tag)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}

	writer.file = f
	return nil
}

func (writer *FileWriter) Close() error {
	return writer.file.Close()
}

func (writer *FileWriter) Write(b []byte) (n int, err error) {
	if len(b) > writer.config.fileBufferSize {
		// Break up messages that exceed the downstream buffer length,
		// using a bytes.Buffer since it's easy. This may result in an
		// undesirable amount of allocations, but the assumption is that
		// bursts of too-long messages are rare.
		buf := bytes.NewBuffer(b)
		for buf.Len() > 0 {
			writer.file.Write(buf.Next(writer.config.fileBufferSize))
			writer.file.Write([]byte(NL))
		}
	} else {
		writer.file.Write(b)
		writer.file.Write([]byte(NL))
	}

	return len(b), nil
}
