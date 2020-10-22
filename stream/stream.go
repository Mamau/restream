package stream

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const RTMP_ADDRESS = "rtmp://0.0.0.0:1935/stream/"

type Stream struct {
	isStarted          bool
	FileName           string `json:"filename"`
	Name               string `json:"name"`
	channelCommand     chan *exec.Cmd
	stopChannelCommand chan bool
	streamDuration     time.Duration
}

func InitStream() Stream {
	return Stream{
		channelCommand:     make(chan *exec.Cmd),
		stopChannelCommand: make(chan bool),
		streamDuration:     10 * time.Minute,
	}
}

func (s *Stream) Start() {
	if s.isStarted {
		return
	}
	cmd := exec.Command("ffmpeg", "-i", s.FileName, "-c", "copy", "-f", "flv", s.getStreamAddress())
	go s.startCommandAtChannel(cmd)
	go s.receiveChannelData()
}

func (s *Stream) Stop() {
	if s.isStarted {
		s.stopChannelCommand <- true
	}
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
			s.killStream()
			return
		}
	}
}

func (s *Stream) killStream() {
	command := <-s.channelCommand
	fmt.Println("Kill process with PID: ", command.Process.Pid)
	err := command.Process.Kill()
	if err != nil {
		log.Fatalf("Cant kill process %v, error: %v", command.Process.Pid, err)
	}

	_, errWait := command.Process.Wait()
	if errWait != nil {
		log.Fatalf("Cant wait process %v, error: %v", command.Process.Pid, errWait)
	}
	s.isStarted = false
}

func (s *Stream) startCommandAtChannel(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Cant start rtmp stream %s\n", err)
	}
	s.isStarted = true
	s.channelCommand <- cmd
}
