package stream

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"time"
)

type ScheduledStream struct {
	*Stream
	startChannelCommand chan bool
	StartAt             int64 `json:"startAt"`
	StopAt              int64 `json:"stopAt"`
}

func NewScheduledStream() *ScheduledStream {
	return &ScheduledStream{
		Stream:              NewStream(),
		startChannelCommand: make(chan bool),
	}
}

func (s *ScheduledStream) ScheduleDownload() {
	if s.isStarted {
		return
	}

	t := time.Now()
	startAfter := s.StartAt - t.Unix()
	stopAfter := s.StopAt - t.Unix()
	format := "15:04:05_02.01.2006"
	pwd, err := os.Getwd()
	if err != nil {
		zap.L().Error("cant get current directory")
		return
	}

	s.SetContext(time.Duration(stopAfter) * time.Second)
	folder := fmt.Sprintf("%v/%v-%v", pwd, s.StartAt, s.StopAt)
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		zap.L().Error("cant create folder",
			zap.String("folder", folder),
			zap.String("error", err.Error()),
		)
		return
	}

	s.logPath, err = os.Create(fmt.Sprintf("%v/log.txt", folder))
	if err != nil {
		panic(err)
	}

	s.Name = fmt.Sprintf("%v/%v", folder, s.Name)

	if err := GetLive().SetStream(s); err != nil {
		zap.L().Error("cant schedule stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		return
	}

	zap.L().Info("stream scheduled download",
		zap.String("startAfter", time.Unix(s.StartAt, 10).Format(format)),
		zap.String("stopAfter", time.Unix(s.StopAt, 10).Format(format)),
	)
	time.AfterFunc(time.Duration(startAfter)*time.Second, s.Download)
}
