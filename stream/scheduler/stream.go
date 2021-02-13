package scheduler

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/stream"
	"github.com/mamau/restream/stream/mpeg"
	"net/url"
	"os"
	"strings"
	"time"
)

type ScheduledStream struct {
	*stream.Stream
	startChannelCommand chan bool
	StartAt             int64 `json:"startAt"`
	StopAt              int64 `json:"stopAt"`
}

func NewScheduledStream() *ScheduledStream {
	return &ScheduledStream{
		Stream:              stream.NewStream(),
		startChannelCommand: make(chan bool),
	}
}

func (s *ScheduledStream) ScheduleDownload() error {
	if s.IsStarted {
		return errors.New(fmt.Sprintf("stream %v already started\n", s.Name))
	}

	t := time.Now()
	startAfter := s.StartAt - t.Unix()
	format := "15:04:05_02.01.2006"
	formatFolder := "15_04_05"

	pwd, err := os.Getwd()
	if err != nil {
		return errors.New("cant get current directory")
	}
	st := time.Unix(s.StartAt, 10).Format(formatFolder)
	sp := time.Unix(s.StopAt, 10).Format(formatFolder)

	if err := stream.GetLive().SetStream(s); err != nil {
		return err
	}

	folder := fmt.Sprintf("%v/%v_%v-%v", pwd, s.Name, st, sp)
	var downloader stream.Downloader
	if strings.Contains(s.Manifest, ".mpd") {
		url4eg, err := url.Parse(s.Manifest)
		if err != nil {
			s.Logger.Fatal(err)
		}
		downloader = mpeg.NewMpegDash(s.Name, folder, url4eg)
	} else {
		downloader = stream.NewM3u8(s.Name, folder, s.Manifest)
	}
	downloader.SetDeadline(s.StopAt)

	s.Logger.Info("stream scheduled download, startAfter: %s stopAfter: %s \n",
		time.Unix(s.StartAt, 10).Format(format), time.Unix(s.StopAt, 10).Format(format))

	time.AfterFunc(time.Duration(startAfter)*time.Second, func() { s.Download(downloader) })
	return nil
}
