package stream

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium"
	"github.com/mamau/restream/stream/selenium/channel"
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
	s := &Stream{}
	s.setLogger()
	return s
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) SetDeadline(deadLine *time.Time) {
	s.deadLine = deadLine
}

func (s *Stream) setLogger() {
	folder := fmt.Sprintf("%v/%v", helpers.CurrentDir(), "storage/logs")
	s.Logger = storage.NewStreamLogger(folder, s.Name)
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
		s.Logger.FatalLogger.Printf("cant fetch manifest via selenium %s\n", s.Name)
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
	s.Logger.ErrorLogger.Printf("stream %s is not started o.0", s.Name)
}

func (s *Stream) Download(d Downloader) {
	if s.IsStarted {
		return
	}
	s.IsStarted = true
	d.Start()
	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		s.Logger.ErrorLogger.Printf("cant delete stream, stream %v, error %v\n", s.Name, err.Error())
		return
	}
}

func (s *Stream) runCommand(c []string) {
	s.Logger.InfoLogger.Printf("starting, stream %s\n", s.Name)
	s.command = exec.Command("ffmpeg", c...)
	if err := s.command.Start(); err != nil {
		s.Logger.ErrorLogger.Printf("cant start download stream, stream %v, error %v\n", s.Name, err.Error())
		s.stopCommand()
		return
	}
	s.IsStarted = true
}

func (s *Stream) stopCommand() {
	s.Logger.InfoLogger.Printf("stopping command..., PID %v, stream %v cmd %s\n", s.command.Process.Pid, s.Name, s.command.String())

	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		s.Logger.ErrorLogger.Printf("cant delete stream, stream %v cmd %s\n", s.Name, err.Error())
		return
	}

	if err := s.command.Process.Signal(syscall.SIGINT); err != nil {
		s.Logger.ErrorLogger.Printf("cant stop process, stream %v, PID %v cmd %s\n", s.Name, s.command.Process.Pid, err.Error())
		s.Logger.InfoLogger.Println("lets kill it")
		if err := s.command.Process.Kill(); err != nil {
			s.Logger.FatalLogger.Printf("cant kill process, stream %v, PID %v cmd %s\n", s.Name, s.command.Process.Pid, err.Error())
		}
		return
	}

	if _, err := s.command.Process.Wait(); err != nil {
		s.Logger.FatalLogger.Printf("cant wait process, stream %v, PID %v cmd %s\n", s.Name, s.command.Process.Pid, err.Error())
	}

	s.Logger.InfoLogger.Printf("command stopped stream %s\n", s.Name)
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(RTMP_ADDRESS)
	address.WriteString(s.Name)
	return address.String()
}
