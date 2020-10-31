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
	}
}

func (s *ScheduledStream) scheduleCmd() {
	t := time.Now()
	start := s.StartAt - t.Unix()
	stop := s.StopAt - t.Unix()
	zap.L().Info("stream scheduled download",
		zap.Int("start", int(start)),
		zap.Int("stop", int(stop)),
	)

	s.Name = fmt.Sprintf("%v_start_at_%v_stop_at_%v.mp4", s.Name, start, stop)
	time.AfterFunc(time.Duration(start)*time.Second, s.Download)
	time.AfterFunc(time.Duration(stop)*time.Second, s.killStream)
}
