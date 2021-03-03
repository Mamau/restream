package stream

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium"
	"github.com/mamau/restream/stream/selenium/channel"
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
	afterDeadline *time.Timer
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) Start() bool {
	if s.IsStarted {
		storage.GetLogger().Warning(fmt.Sprintf("stream %v already started\n", s.Name))
		return false
	}

	if s.Manifest == "" {
		storage.GetLogger().Error(errors.New(fmt.Sprintf("no manifest file at stream: %s \n", s.Name)))
		return false
	}

	if err := GetLive().SetStream(s); err != nil {
		storage.GetLogger().Warning(err.Error())
		return false
	}

	if isTest, _ := strconv.ParseBool(os.Getenv("IS_TEST")); isTest {
		s.IsStarted = s.runTestCommand()
	} else {
		s.IsStarted = s.runCommand([]string{"-re", "-i", s.Manifest, "-acodec", "copy", "-vcodec", "copy", "-f", "flv", s.getStreamAddress()})
	}

	return s.IsStarted
}

func (s *Stream) StartWithDeadLine() bool {
	if s.IsStarted {
		storage.GetLogger().Warning(fmt.Sprintf("stream %v already started\n", s.Name))
		return false
	}

	if s.DeadLine == nil {
		storage.GetLogger().Info("no deadline... set default for 04:00:00")
		format := "15:04:05"
		now := time.Now()
		stop, err := time.Parse(format, "04:00:00")
		if err != nil {
			storage.GetLogger().Fatal(err)
		}

		if now.After(stop) {
			now = now.Add(time.Hour * 24)
		}
		stop = time.Date(now.Year(), now.Month(), now.Day(), stop.Hour(), stop.Minute(), stop.Second(), 0, time.Local)
		s.DeadLine = &stop
	}

	isStated := s.Start()

	if isStated {
		//end := s.DeadLine.Unix() - time.Now().Unix()
		//
		//duration := time.Duration(end) * time.Second
		//s.afterDeadline = time.AfterFunc(duration, s.Stop)
		go s.stopAfterDuration()
	}

	return isStated
}

func (s *Stream) stopAfterDuration() {
	end := s.DeadLine.Unix() - time.Now().Unix()
	duration := time.Duration(end) * time.Second
	hasCome := make(chan bool)

	time.AfterFunc(duration, func() {
		hasCome <- true
	})

	for {
		select {
		case <-hasCome:
			s.Stop()
			return
		}
	}
}

func (s *Stream) StartViaSelenium(withDeadline bool) bool {
	manifest, err := selenium.GetManifest(channel.Channel(s.Name))
	if err != nil {
		storage.GetLogger().Fatal(errors.New(fmt.Sprintf("cant fetch manifest via selenium %s, err: %s\n", s.Name, err.Error())))
	}

	s.Manifest = manifest

	if withDeadline {
		return s.StartWithDeadLine()
	}
	return s.Start()
}

func (s *Stream) Stop() {
	if !s.IsStarted {
		storage.GetLogger().Warning(fmt.Sprintf("stream %s is not started", s.Name))
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
		storage.GetLogger().Error(err)
		return
	}
}
func (s *Stream) runTestCommand() bool {
	s.command = exec.Command("ping", "ya.ru")
	s.command.Stdout = os.Stdout
	s.command.Stderr = os.Stderr
	if err := s.command.Start(); err != nil {
		storage.GetLogger().Error(err)
		s.stopCommand()
		return false
	}
	return true
}

func (s *Stream) runCommand(c []string) bool {
	stopAfterInfo := ""
	if s.DeadLine != nil {
		format := "15:04:05 02.01.2006"
		stopAfterInfo = fmt.Sprintf(", stop after deadline: %v\n", s.DeadLine.Format(format))
	}

	storage.GetLogger().Info("starting, stream %s%s\n", s.Name, stopAfterInfo)
	s.command = exec.Command("ffmpeg", c...)
	if err := s.command.Start(); err != nil {
		storage.GetLogger().Error(err)
		s.stopCommand()
		return false
	}

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

	return true
}

func (s *Stream) isManifestAvailable(t *time.Ticker) {
	if s.Manifest == "" {
		storage.GetLogger().Warning(fmt.Sprintf("empty manifest, on stream %s", s.Name))
	}
	resp, err := http.Get(s.Manifest)
	if err != nil {
		storage.GetLogger().Error(err)
		return
	}

	if isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices; !isOk {
		storage.GetLogger().Warning("manifest is not available %s\n", s.Manifest)
		t.Stop()
		s.Restart()
		return
	}
}

func (s *Stream) Restart() {
	storage.GetLogger().Info("restart stream %s \n", s.Name)
	s.Stop()

	hasDeadline := s.DeadLine != nil

	s.StartViaSelenium(hasDeadline)
}

func (s *Stream) stopCommand() {
	storage.GetLogger().Info("stopping command..., PID %v, stream %v cmd %s\n", s.command.Process.Pid, s.Name, s.command.String())

	if err := s.command.Process.Signal(syscall.SIGINT); err != nil {
		storage.GetLogger().Error(err)
		if err := s.command.Process.Kill(); err != nil {
			storage.GetLogger().Fatal(err)
		}
		return
	}

	if _, err := s.command.Process.Wait(); err != nil {
		storage.GetLogger().Fatal(err)
	}

	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		storage.GetLogger().Error(err)
		return
	}

	s.IsStarted = false
	storage.GetLogger().Info("command stopped stream %s\n", s.Name)
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(os.Getenv("RTMP_ADDRESS"))
	address.WriteString(s.Name)
	return address.String()
}
