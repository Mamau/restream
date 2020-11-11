package stream

import (
	"bytes"
	"context"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"syscall"
	"time"
)

const RTMP_ADDRESS = "rtmp://0.0.0.0:1935/stream/"

type Streamer interface {
	GetName() string
	Start()
	Stop()
	Download()
	SetContext(d time.Duration)
}

type Stream struct {
	isStarted bool
	FileName  string `json:"filename"`
	Name      string `json:"name"`
	quality   string
	command   *exec.Cmd
	ctx       context.Context
}

func NewStream() *Stream {
	s := Stream{
		quality: "p:1",
	}
	s.SetContext(1 * time.Hour)
	return &s
}

func (s *Stream) SetContext(d time.Duration) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(d))
	defer cancel()
	s.ctx = ctx

	select {
	case <-s.ctx.Done():
		s.stopCommand()
		return
	}
}

func (s *Stream) GetName() string {
	return s.Name
}

func (s *Stream) Start() {
	if s.isStarted {
		return
	}
	go s.setCommand([]string{"-loglevel", "verbose", "-re", "-i", s.FileName, "-vcodec", "libx264", "-vprofile", "baseline", "-acodec", "libmp3lame", "-ar", "44100", "-ac", "1", "-f", "flv", s.getStreamAddress()})
}

func (s *Stream) Stop() {
	if s.isStarted {
		s.stopCommand()
	}
}

func (s *Stream) Download() {
	if s.isStarted {
		return
	}

	outputFile := fmt.Sprintf("%v.mp4", s.Name)
	go s.setCommand([]string{"-i", s.FileName, "-map", s.quality, "-c", "copy", "-bsf:a", "aac_adtstoasc", outputFile})
}

func (s *Stream) setCommand(c []string) {
	s.command = exec.CommandContext(s.ctx, "ffmpeg", c...)
	err := s.command.Start()
	if err != nil {
		zap.L().Error("cant start download stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
		s.stopCommand()
	}
	s.isStarted = true
}

func (s *Stream) stopCommand() {
	zap.L().Info("stopping command...", zap.String("stream", s.Name), zap.String("cmd", s.command.String()))
	if _, err := GetLive().DeleteStream(s.Name); err != nil {
		zap.L().Error("cant delete stream",
			zap.String("stream", s.Name),
			zap.String("error", err.Error()),
		)
	}

	if err := s.command.Process.Signal(syscall.SIGTERM); err != nil {
		zap.L().Error("cant stop process",
			zap.String("stream", s.Name),
			zap.Int("PID", s.command.Process.Pid),
			zap.String("error", err.Error()),
		)
	}
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(RTMP_ADDRESS)
	address.WriteString(s.Name)
	return address.String()
}
