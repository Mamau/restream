package storage

import (
	"fmt"
	"log"
	"os"
)

type LogType string

const (
	INFO  LogType = "info"
	ERROR LogType = "err"
)

type StreamLogger struct {
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	FatalLogger   *log.Logger
}

func NewStreamLogger(folder, name string) *StreamLogger {
	s := StreamLogger{}
	s.initLogger(folder, name)
	return &s
}

func (s *StreamLogger) initLogger(folder, name string) {
	fullFolderName := fmt.Sprintf("%s/%s", folder, name)
	if err := os.MkdirAll(fullFolderName, os.ModePerm); err != nil {
		log.Fatalf("cant create folder %s\n", fullFolderName)
	}

	folders := map[LogType]string{
		INFO:  fmt.Sprintf("%s/%s", fullFolderName, INFO),
		ERROR: fmt.Sprintf("%s/%s", fullFolderName, ERROR),
	}

	createdFolders := map[LogType]*os.File{}

	for k, v := range folders {
		file, err := os.OpenFile(fmt.Sprintf("%s_log.txt", v), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		createdFolders[k] = file
	}

	s.InfoLogger = log.New(createdFolders[INFO], "INFO: ", log.Ldate|log.Ltime)
	s.WarningLogger = log.New(createdFolders[ERROR], "WARNING: ", log.Ldate|log.Ltime)
	s.ErrorLogger = log.New(createdFolders[ERROR], "ERROR: ", log.Ldate|log.Ltime)
	s.FatalLogger = log.New(createdFolders[ERROR], "FATAL: ", log.Ldate|log.Ltime)
}
