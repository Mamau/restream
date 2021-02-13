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

const RTMP_ADDRESS = "rtmp://nginx-rtmp:1935/stream/"

//const RTMP_ADDRESS = "rtmp://0.0.0.0:1935/stream/"

type Streamer interface {
	GetName() string
	Start() error
	Stop()
	Download(Downloader)
}

type Downloader interface {
	Start()
	Stop()
	SetDeadline(stopAt int64)
}

type Stream struct {
	IsStarted bool
	Manifest  string `json:"manifest"`
	Name      string `json:"name"`
	logPath   *os.File
	deadLine  *time.Time
	command   *exec.Cmd
	Logger    *storage.StreamLogger
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

func (s *Stream) SetDeadline(deadLine *time.Time) {
	s.deadLine = deadLine
}

func (s *Stream) Start() error {
	if s.IsStarted {
		return errors.New(fmt.Sprintf("stream %v already started\n", s.Name))
	}

	if s.Manifest == "" {
		return errors.New(fmt.Sprintf("no manifest file at stream: %s \n", s.Name))
	}

	if err := GetLive().SetStream(s); err != nil {
		return err
	}

	go s.runCommand([]string{"-re", "-i", s.Manifest, "-acodec", "copy", "-vcodec", "copy", "-f", "flv", s.getStreamAddress()})
	return nil
}

func (s *Stream) StartWithDeadLine() error {
	if s.deadLine == nil {
		return errors.New(fmt.Sprintf("need set deadline"))
	}
	if err := s.Start(); err != nil {
		return errors.New(fmt.Sprintf("cant start stream %s with deadline, err: %s\n", s.Name, err.Error()))
	}

	end := s.deadLine.Unix() - time.Now().Unix()
	time.AfterFunc(time.Duration(end)*time.Second, func() {
		s.Stop()
	})
	return nil
}

func (s *Stream) StartViaSelenium(withDeadline bool) error {
	if err := s.fetchManifest(); err != nil {
		s.Logger.Warning("cant fetch manifest via selenium %s\n", s.Name)
	}

	if withDeadline {
		return s.StartWithDeadLine()
	}
	return s.Start()
}

func (s *Stream) fetchManifest() error {
	manifest, err := selenium.GetManifest(channel.Channel(s.Name))
	if err != nil {
		return err
	}
	s.Manifest = manifest
	return nil
}

func (s *Stream) Stop() {
	if s.IsStarted {
		s.stopCommand()
	}
	s.Logger.Warning("stopped stream %s is not started o.0", s.Name)
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
		s.Logger.Warning("-----manifest is not available %s -----\n", s.Manifest)
		t.Stop()
		s.Restart()
		return
	}
}

func (s *Stream) Restart() {
	s.Logger.Info("restart stream %s \n", s.Name)
	s.Stop()
	if err := s.StartViaSelenium(true); err != nil {
		s.Logger.Fatal(err)
	}
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

	s.Logger.Info("command stopped stream %s\n", s.Name)
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(RTMP_ADDRESS)
	address.WriteString(s.Name)
	return address.String()
}
