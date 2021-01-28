package stream

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/stream/mpeg"
	"go.uber.org/zap"
	"log"
	"net/url"
	"os"
	"strings"
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

func (s *ScheduledStream) ScheduleDownload() error {
	if s.isStarted {
		return errors.New(fmt.Sprintf("stream %v already started\n", s.Name))
	}

	t := time.Now()
	startAfter := s.StartAt - t.Unix()
	format := "15:04:05_02.01.2006"
	formatFolder := "15_04_05"

	pwd, err := os.Getwd()
	if err != nil {
		zap.L().Error("cant get current directory")
		return errors.New("cant get current directory")
	}
	st := time.Unix(s.StartAt, 10).Format(formatFolder)
	sp := time.Unix(s.StopAt, 10).Format(formatFolder)

	if err := GetLive().SetStream(s); err != nil {
		zap.L().Error("cant schedule stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		return err
	}

	folder := fmt.Sprintf("%v/%v_%v-%v", pwd, s.Name, st, sp)
	var downloader Downloader
	if strings.Contains(s.Manifest, ".mpd") {
		url4eg, err := url.Parse(s.Manifest)
		if err != nil {
			log.Fatalf("err creating url %v\n", err)
		}
		downloader = mpeg.NewMpegDash(s.Name, folder, url4eg)
	} else {
		downloader = NewM3u8(s.Name, folder, s.Manifest)
	}
	downloader.SetDeadline(s.StopAt)

	zap.L().Info("stream scheduled download",
		zap.String("startAfter", time.Unix(s.StartAt, 10).Format(format)),
		zap.String("stopAfter", time.Unix(s.StopAt, 10).Format(format)),
	)

	time.AfterFunc(time.Duration(startAfter)*time.Second, func() { s.Download(downloader) })
	return nil
}
