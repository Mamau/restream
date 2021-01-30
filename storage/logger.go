package storage

import (
	"fmt"
	"github.com/mamau/restream/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

func InitLogger() {
	cf := zap.NewDevelopmentConfig()
	cf.OutputPaths = []string{"stdout", createLogFile()}
	cf.Encoding = "json"
	cf.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:     "time",
		EncodeTime:  zapcore.TimeEncoderOfLayout("02.01.2006 15:04:05"),
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
	}
	logger, err := cf.Build()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	zap.ReplaceGlobals(logger)
}

func createLogFile() string {
	folder := fmt.Sprintf("%v/%v", helpers.CurrentDir(), "logs")
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Fatalf("cant create folder %s\n", folder)
	}

	filePath := fmt.Sprintf("%v/stream.log", folder)
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("create logFile error: %v\n", err)
	}
	if err := logFile.Close(); err != nil {
		log.Fatalf("cant close fole %s\n", filePath)
	}

	return filePath
}
