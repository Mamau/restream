package storage

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() {
	cf := zap.NewDevelopmentConfig()
	cf.OutputPaths = []string{"stdout", "/tmp/stream.log"}
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
