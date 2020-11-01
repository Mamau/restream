package stream

import (
	"fmt"
	"go.uber.org/zap"
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
	if !s.isStarted {
		s.scheduleCmd()
		if err := GetLive().ScheduleStream(s); err != nil {
			zap.L().Fatal("cant schedule stream",
				zap.String("stream", s.Name),
				zap.String("error", err.Error()),
			)
		}
	}
}

func (s *ScheduledStream) scheduleCmd() {
	t := time.Now()
	startAfter := s.StartAt - t.Unix()
	stopAfter := s.StopAt - t.Unix()
	format := "15_04_05_02_01_2006"

	//killStreamWithDelay := stopAfter + 10
	startAt := time.Unix(s.StartAt, 10).Format(format)
	stopAt := time.Unix(s.StopAt, 10).Format(format)

	zap.L().Info("stream scheduled download",
		zap.String("startAfter", startAt),
		zap.String("stopAfter", stopAt),
	)

	s.Name = fmt.Sprintf("%v-%v-%v", startAt, stopAt, s.Name)

	s.streamDuration = time.Duration(stopAfter) * time.Second
	//s.streamDuration = 30 * time.Second

	//s.Download()
	time.AfterFunc(time.Duration(startAfter)*time.Second, s.Download)
	//time.AfterFunc(time.Duration(killStreamWithDelay)*time.Second, s.killStream)
}
