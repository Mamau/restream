package stream

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium"
	"github.com/mamau/restream/stream/selenium/channel"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Streamer interface {
	GetName() string
	Start() bool
	Stop()
	Download(Downloader)
	Restart()
}

type Downloader interface {
	Start()
	Stop()
	SetDeadline(stopAt int64)
}

type Stream struct {
	IsStarted     bool
	Manifest      string `json:"manifest"`
	Name          string `json:"name"`
	logPath       *os.File
	DeadLine      *time.Time
	command       *exec.Cmd
	Logger        *storage.StreamLogger
	afterDeadline *time.Timer
}

func NewStream() *Stream {
	s := &Stream{
		Logger: storage.NewStreamLogger(),
	}
	return s
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) Start() bool {
	if s.IsStarted {
		s.Logger.Warning(fmt.Sprintf("stream %v already started\n", s.Name))
		return false
	}

	if s.Manifest == "" {
		s.Logger.Error(errors.New(fmt.Sprintf("no manifest file at stream: %s \n", s.Name)))
		return false
	}

	if err := GetLive().SetStream(s); err != nil {
		s.Logger.Error(err)
		return false
	}

	go s.runCommand([]string{"-re", "-i", s.Manifest, "-acodec", "copy", "-vcodec", "copy", "-f", "flv", s.getStreamAddress()})
	return true
}

func (s *Stream) StartWithDeadLine() bool {
	if s.IsStarted {
		s.Logger.Warning(fmt.Sprintf("stream %v already started\n", s.Name))
		return false
	}

	if s.DeadLine == nil {
		s.Logger.Fatal(errors.New(fmt.Sprintf("need set deadline")))
		return false
	}

	isStated := s.Start()

	if isStated {
		if s.DeadLine.Unix() < time.Now().Unix() {
			s.Logger.Warning("why u have deadline less than now time?")
		}

		end := s.DeadLine.Unix() - time.Now().Unix()

		duration := time.Duration(end) * time.Second
		format := "15:04:05 02.01.2006"
		s.Logger.Info(fmt.Sprintf("stop after %v, deadline :%v, now: %v\n", duration, s.DeadLine.Format(format), time.Now().Format(format)))

		s.afterDeadline = time.AfterFunc(duration, s.Stop)
	}

	return isStated
}

func (s *Stream) StartViaSelenium(withDeadline bool) bool {
	manifest, err := selenium.GetManifest(channel.Channel(s.Name))
	if err != nil {
		s.Logger.Fatal(errors.New(fmt.Sprintf("cant fetch manifest via selenium %s, err: %s\n", s.Name, err.Error())))
	}

	s.Manifest = manifest

	if withDeadline {
		return s.StartWithDeadLine()
	}
	return s.Start()
}

func (s *Stream) Stop() {
	if !s.IsStarted {
		s.Logger.Warning("stopped stream %s is not started o.0", s.Name)
		return
	}
	s.stopCommand()
}

func (s *Stream) Download(d Downloader) {
	if s.IsStarted {
		return
	}
	s.IsStarted = true
	d.Start()
	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		s.Logger.Error(err)
		return
	}
}

func (s *Stream) runCommand(c []string) {
	s.Logger.Info("starting, stream %s\n", s.Name)
	s.command = exec.Command("ffmpeg", c...)
	if err := s.command.Start(); err != nil {
		s.Logger.Error(err)
		s.stopCommand()
		return
	}
	s.IsStarted = true

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if s.IsStarted == false {
					ticker.Stop()
					return
				}
				s.isManifestAvailable(ticker)
			}
		}
	}()
}

func (s *Stream) isManifestAvailable(t *time.Ticker) {
	resp, err := http.Get(s.Manifest)
	if err != nil {
		s.Logger.Error(err)
		return
	}

	if isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices; !isOk {
		s.Logger.Warning("manifest is not available %s\n", s.Manifest)
		t.Stop()
		s.Restart()
		return
	}
}

func (s *Stream) Restart() {
	s.Logger.Info("restart stream %s \n", s.Name)
	s.Stop()

	hasDeadline := false
	if s.DeadLine != nil {
		t := time.Now()
		if s.DeadLine.Before(time.Now()) {
			t = t.Add(time.Hour * 24)
		}

		stop := time.Date(t.Year(), t.Month(), t.Day(), s.DeadLine.Hour(), s.DeadLine.Minute(), s.DeadLine.Second(), 0, time.UTC)
		s.Logger.Info("set new deadline %v", stop.Unix())
		s.DeadLine = &stop
		s.afterDeadline.Reset(time.Duration(stop.Unix()) * time.Second)
		hasDeadline = true
	}

	s.StartViaSelenium(hasDeadline)
}

func (s *Stream) stopCommand() {
	s.Logger.Info("stopping command..., PID %v, stream %v cmd %s\n", s.command.Process.Pid, s.Name, s.command.String())

	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		s.Logger.Error(err)
		return
	}

	if err := s.command.Process.Signal(syscall.SIGINT); err != nil {
		s.Logger.Error(err)
		if err := s.command.Process.Kill(); err != nil {
			s.Logger.Fatal(err)
		}
		return
	}

	if _, err := s.command.Process.Wait(); err != nil {
		s.Logger.Fatal(err)
	}

	s.IsStarted = false
	s.Logger.Info("command stopped stream %s\n", s.Name)
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(os.Getenv("RTMP_ADDRESS"))
	address.WriteString(s.Name)
	return address.String()
}
