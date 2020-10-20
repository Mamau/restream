package stream

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

type Stream struct {
	isStarted          bool
	fileName           string
	rtmpAddress        string
	channelCommand     chan *exec.Cmd
	stopChannelCommand chan bool
	streamDuration     time.Duration
}

func InitStream() Stream {
	return Stream{
		fileName:           "https://matchtv.ru/vdl/playlist/133529/adaptive/1603241852/e00f0f847bab9f4bf0faef8eed6666ff/web.m3u8",
		rtmpAddress:        "rtmp://0.0.0.0:1935/stream/mystream",
		channelCommand:     make(chan *exec.Cmd),
		stopChannelCommand: make(chan bool),
		streamDuration:     1 * time.Minute,
	}
}

func (s *Stream) Start(queryValues url.Values) {
	if s.isStarted || !s.isFileManifestAvailable() || !s.isNginxRestreamRunning() {
		return
	}
	cmd := s.prepareStreamCmd(queryValues)
	go s.startCommandAtChannel(cmd)
	go s.receiveChannelData()
}

func (s *Stream) Stop() {
	if s.isStarted {
		s.stopChannelCommand <- true
	}
}

func (s *Stream) prepareStreamCmd(queryValues url.Values) *exec.Cmd {
	if queryValues.Get("file") != "" {
		s.fileName = queryValues.Get("file")
	}

	return exec.Command("ffmpeg", "-i", s.fileName, "-c", "copy", "-f", "flv", s.rtmpAddress)
}

func (s *Stream) isFileManifestAvailable() bool {
	resp, err := http.Get(s.fileName)
	if err != nil {
		log.Fatalf("Error while check manifest %v with err: %v\n", s.fileName, err)
	}
	isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !isOk {
		fmt.Printf("File %v is not available %v\n", s.fileName, resp.StatusCode)
	}
	return isOk
}

func (s *Stream) isNginxRestreamRunning() bool {
	resp, err := http.Get("http://0.0.0.0:8081")
	if err != nil {
		fmt.Printf("Nginx is not running, error: %v\n", err)
		return false
	}

	isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !isOk {
		fmt.Printf("Nginx restream %v is not available %v\n", s.rtmpAddress, resp.StatusCode)
	}
	return isOk
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
