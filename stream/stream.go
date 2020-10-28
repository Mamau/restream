package stream

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

const RTMP_ADDRESS = "rtmp://0.0.0.0:1935/stream/"

type Stream struct {
	isStarted          bool
	FileName           string `json:"filename"`
	Name               string `json:"name"`
	quality            string
	channelCommand     chan *exec.Cmd
	stopChannelCommand chan bool
	streamDuration     time.Duration
}

func NewStream() *Stream {
	return &Stream{
		quality:            "p:1",
		channelCommand:     make(chan *exec.Cmd),
		stopChannelCommand: make(chan bool),
		streamDuration:     10 * time.Minute,
	}
}

func (s *Stream) Start() {
	if s.isStarted {
		return
	}
	//cmd := exec.Command("ffmpeg", "-i", s.FileName, "-c", "copy", "-f", "flv", s.getStreamAddress())
	cmd := exec.Command("ffmpeg", "-loglevel", "verbose", "-re", "-i", s.FileName, "-vcodec", "libx264", "-vprofile", "baseline", "-acodec", "libmp3lame", "-ar", "44100", "-ac", "1", "-f", "flv", s.getStreamAddress())
	//cmd := exec.Command("ping", "ya.ru")
	s.startFfmpegCmd(cmd)
}

func (s *Stream) Stop() {
	if s.isStarted {
		s.stopChannelCommand <- true
	}
}

func (s *Stream) Download() {
	if s.isStarted {
		return
	}
	outpuFile := fmt.Sprintf("%v.mp4", s.Name)
	cmd := exec.Command("ffmpeg", "-i", s.FileName, "-map", s.quality, "-c", "copy", "-bsf:a", "aac_adtstoasc", outpuFile)
	s.startFfmpegCmd(cmd)
}

func (s *Stream) startFfmpegCmd(cmd *exec.Cmd) {
	go s.startCommandAtChannel(cmd)
	go s.receiveChannelData()
}

func (s *Stream) getStreamAddress() string {
	address := bytes.Buffer{}
	address.WriteString(RTMP_ADDRESS)
	address.WriteString(s.Name)
	return address.String()
}

func (s *Stream) receiveChannelData() {
	for {
		select {
		case isStopChannel := <-s.stopChannelCommand:
			if isStopChannel {
				s.killStream()
				return
			}
		case <-time.After(s.streamDuration):
			zap.L().Info("kill after time",
				zap.Int("duration", int(s.streamDuration)),
			)
			s.killStream()
			return
		}
	}
}

func (s *Stream) killStream() {
	command := <-s.channelCommand
	zap.L().Info("kill stream",
		zap.String("stream", s.Name),
		zap.Int("PID", command.Process.Pid),
	)
	err := command.Process.Kill()
	if err != nil {
		zap.L().Fatal("cant kill process",
			zap.String("stream", s.Name),
			zap.Int("PID", command.Process.Pid),
			zap.String("error", err.Error()),
		)
	}

	_, errWait := command.Process.Wait()
	if errWait != nil {
		zap.L().Fatal("cant wait process",
			zap.String("stream", s.Name),
			zap.Int("PID", command.Process.Pid),
			zap.String("error", errWait.Error()),
		)
	}
	s.isStarted = false
}

func (s *Stream) startCommandAtChannel(cmd *exec.Cmd) {
	zap.L().Info("start command",
		zap.String("stream", s.Name),
		zap.String("cmd", cmd.String()),
	)

	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		zap.L().Fatal("failed start command:",
			zap.String("stream", s.Name),
			zap.String("cmd", cmd.String()),
			zap.String("error", err.Error()),
		)
	}
	s.isStarted = true
	s.channelCommand <- cmd
}
