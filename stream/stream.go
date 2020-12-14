package stream

import (
	"bytes"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"syscall"
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
	isStarted bool
	FileName  string `json:"filename"`
	Name      string `json:"name"`
	logPath   *os.File
	command   *exec.Cmd
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) Start() error {
	if s.isStarted {
		return errors.New(fmt.Sprintf("stream %v already started\n", s.Name))
	}
	if err := GetLive().SetStream(s); err != nil {
		return err
	}

	go s.runCommand([]string{"-re", "-i", s.FileName, "-acodec", "copy", "-vcodec", "copy", "-f", "flv", s.getStreamAddress()})
	return nil
}

func (s *Stream) Stop() {
	if s.isStarted {
		s.stopCommand()
	}
}

func (s *Stream) Download(d Downloader) {
	if s.isStarted {
		return
	}
	s.isStarted = true
	d.Start()
	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		zap.L().Error("cant delete stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		return
	}
}

func (s *Stream) runCommand(c []string) {
	s.command = exec.Command("ffmpeg", c...)
	s.command.Stdout = os.Stdout
	s.command.Stderr = os.Stderr
	if err := s.command.Run(); err != nil {
		zap.L().Error("cant start download stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		s.stopCommand()
		return
	}
	s.isStarted = true
}

func (s *Stream) stopCommand() {
	zap.L().Info("stopping command...",
		zap.Int("PID", s.command.Process.Pid),
		zap.String("stream", s.Name),
		zap.String("cmd", s.command.String()),
	)

	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		zap.L().Error("cant delete stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		return
	}

	if err := s.command.Process.Signal(syscall.SIGINT); err != nil {
		zap.L().Error("cant stop process",
			zap.String("stream", s.Name),
			zap.Int("PID", s.command.Process.Pid),
			zap.String("error", err.Error()),
		)
		zap.L().Info("lets kill it")
		if err := s.command.Process.Kill(); err != nil {
			zap.L().Fatal("cant kill process",
				zap.String("stream", s.Name),
				zap.Int("PID", s.command.Process.Pid),
				zap.String("error", err.Error()),
			)
		}
		return
	}

	if _, err := s.command.Process.Wait(); err != nil {
		zap.L().Fatal("cant wait process",
			zap.String("stream", s.Name),
			zap.Int("PID", s.command.Process.Pid),
			zap.String("error", err.Error()),
		)
	}

	zap.L().Info("command stopped", zap.String("stream", s.Name))
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(RTMP_ADDRESS)
	address.WriteString(s.Name)
	return address.String()
}
