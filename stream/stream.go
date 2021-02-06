package stream

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const RTMP_ADDRESS = "rtmp://nginx-rtmp:1935/stream/"

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
	command   *exec.Cmd
	Logger    *storage.StreamLogger
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) setLogger() {
	folder := fmt.Sprintf("%v/%v", helpers.CurrentDir(), "storage/logs")
	s.Logger = storage.NewStreamLogger(folder, s.Name)
}

func (s *Stream) Start() error {
	s.setLogger()
	if s.IsStarted {
		return errors.New(fmt.Sprintf("stream %v already started\n", s.Name))
	}
	if err := GetLive().SetStream(s); err != nil {
		return err
	}

	go s.runCommand([]string{"-re", "-i", s.Manifest, "-acodec", "copy", "-vcodec", "copy", "-f", "flv", s.getStreamAddress()})
	return nil
}

func (s *Stream) StartWithDeadLine(deadLine time.Time) {
	if err := s.Start(); err != nil {
		s.Logger.ErrorLogger.Printf("cant start stream %s with deadline, err: %s\n", s.Name, err.Error())
	}

	end := deadLine.Unix() - time.Now().Unix()
	time.AfterFunc(time.Duration(end)*time.Second, func() {
		s.Stop()
	})
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
	s.command.Stdout = os.Stdout
	s.command.Stderr = os.Stderr
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
