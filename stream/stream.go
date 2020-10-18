package stream

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"time"
)

type Stream struct {
	fileName           string
	rtmpAddress        string
	channelCommand     chan *exec.Cmd
	stopChannelCommand chan bool
	streamDuration     time.Duration
}

func InitStream() Stream {
	return Stream{
		fileName:           "https://matchtv.ru/vdl/playlist/133529/adaptive/1603070430/f368e523f7b045189de6d704157606d3/web.m3u8",
		rtmpAddress:        "rtmp://0.0.0.0:1935/stream/mystream",
		channelCommand:     make(chan *exec.Cmd),
		stopChannelCommand: make(chan bool),
		streamDuration:     1 * time.Minute,
	}
}

func (s *Stream) Start(queryValues url.Values) {
	cmd := s.prepareStreamCmd(queryValues)

	go s.startCommandAtChannel(cmd)
	go s.receiveChannelData()
}

func (s *Stream) Stop() {
	s.stopChannelCommand <- true
}

func (s *Stream) prepareStreamCmd(queryValues url.Values) *exec.Cmd {
	if queryValues.Get("file") != "" {
		s.fileName = queryValues.Get("file")
	}

	return exec.Command("ffmpeg", "-i", s.fileName, "-c", "copy", "-f", "flv", s.rtmpAddress)
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
	errWait := command.Wait()
	if errWait != nil {
		log.Fatalf("Cant wait process %v, error: %v", command.Process.Pid, err)
	}
	fmt.Println("Closed channel stream")
	close(s.channelCommand)
}

func (s *Stream) startCommandAtChannel(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Cant start rtmp stream %s\n", err)
	}

	s.channelCommand <- cmd
}
