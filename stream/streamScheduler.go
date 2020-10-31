package stream

import (
	"fmt"
	"go.uber.org/zap"
	"strconv"
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

	stopAfterSec := strconv.FormatInt(stop, 10)
	startAt := time.Unix(s.StartAt, 10).Format("15_04_05_02.01.2006")
	stopAt := time.Unix(s.StopAt, 10).Format("15_04_05_02.01.2006")

	zap.L().Info("stream scheduled download",
		zap.String("start", startAt),
		zap.String("stop", stopAt),
	)

	s.Name = fmt.Sprintf("%v_start_at_%v_stop_at_%v", s.Name, startAt, stopAt)
	time.AfterFunc(time.Duration(start)*time.Second, func() {
		s.Download(stopAfterSec)
	})
	//time.AfterFunc(time.Duration(stop)*time.Second, s.killStream)
}
