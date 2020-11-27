package storage

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
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
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		zap.L().Error("cant create folder",
			zap.String("folder", folder),
			zap.String("error", err.Error()),
		)
	}
	file, err := os.OpenFile(fmt.Sprintf("%v/%v_log.txt", folder, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	s.InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	s.WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime)
	s.ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime)
	s.FatalLogger = log.New(file, "FATAL: ", log.Ldate|log.Ltime)
}
