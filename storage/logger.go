package storage

import (
	"github.com/mamau/restream/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

func InitLogger() {
	cf := zap.NewDevelopmentConfig()
	cf.OutputPaths = []string{"stdout", logPath("/storage/logs/info.txt")}
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

func logPath(path string) string {
	var b strings.Builder
	pwd := helpers.Pwd()
	b.WriteString(pwd)
	b.WriteString(path)
	return b.String()
}
