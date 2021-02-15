package mpeg

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"github.com/unki2aut/go-mpd"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type MpegDash struct {
	name        string
	folder      string
	initAudio   bool
	initVideo   bool
	processStop chan bool
	stopped     bool
	audioOrder  int
	videoOrder  int
	chunksAudio []string
	chunksVideo []string
	counter     *helpers.WriteCounter
	media       *media
	manifest    *mpd.MPD
	manifestUrl *url.URL
	deadline    time.Duration
	fileAudio   *os.File
	fileVideo   *os.File
	logger      *storage.StreamLogger
}

func NewMpegDash(name, folder string, manifestUrl *url.URL) *MpegDash {
	m := &MpegDash{
		media: &media{
			chunks: make(map[fileType]chunkMpeg),
		},
		counter:     &helpers.WriteCounter{},
		name:        name,
		manifestUrl: manifestUrl,
		folder:      folder,
		processStop: make(chan bool),
		logger:      storage.NewStreamLogger(),
	}
	m.deadlineJob()
	return m
}
func (m *MpegDash) Start() {
	m.setFiles()
	m.fetchManifest()

	for {
		select {
		case <-m.processStop:
			return
		case <-time.Tick(1 * time.Second):
			if !m.stopped {
				m.fetchChunk(video, m.fileVideo)
				m.fetchChunk(audio, m.fileAudio)
			}
		}
	}
}
func (m *MpegDash) fetchChunk(f fileType, fl *os.File) {
	currentMediaChunk := m.media.GetMediaByType(f)
	if _, ok := helpers.Find(m.chunksVideo, currentMediaChunk); !ok {
		urlVideo := m.getBaseUrl() + currentMediaChunk
		if err := m.downloadFilePart(urlVideo, fl); err != nil {
			m.media.DecrementByType(f)
		} else {
			fmt.Println("Get chunk:", currentMediaChunk, "...")
			m.chunksVideo = m.collectorChunks(m.chunksVideo, currentMediaChunk)
		}
	}
}
func (m *MpegDash) Stop() {
	m.logger.Info("stop mpeg dash")
	m.stopped = true
	time.AfterFunc(7*time.Second, m.closeFiles)
	time.AfterFunc(10*time.Second, m.mergeAudioVideo)
	m.processStop <- true
}
func (m *MpegDash) SetDeadline(stopAt int64) {
	if stopAt <= time.Now().Unix() {
		m.logger.Fatal(errors.New("deadline should be greater than now"))
	}
	m.deadline = time.Duration(stopAt-time.Now().Unix()) * time.Second
	m.deadlineJob()
}
func (m *MpegDash) deadlineJob() {
	if m.deadline == 0 {
		return
	}
	time.AfterFunc(m.deadline, m.Stop)
}
func (m *MpegDash) mergeAudioVideo() {
	m.logger.Info("start merge video %v and audio %v\n", m.fileVideo.Name(), m.fileAudio.Name())
	resultedFileName := fmt.Sprintf("%v/%v.mp4", m.folder, m.name)
	m.logger.Info("total download: %s\n", helpers.ByteHuman(m.counter.GetTotal(), 2))
	command := exec.Command("ffmpeg", "-i", m.fileVideo.Name(), "-i", m.fileAudio.Name(), "-c", "copy", resultedFileName)
	if err := command.Start(); err != nil {
		m.logger.Error(err)
	}
}
func (m *MpegDash) setFiles() {
	if err := os.MkdirAll(m.folder, os.ModePerm); err != nil {
		m.logger.Error(err)
	}

	fileAudio, err := os.OpenFile(fmt.Sprintf("%v/%v_audio.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		m.logger.Fatal(err)
	}
	m.fileAudio = fileAudio

	fileVideo, err := os.OpenFile(fmt.Sprintf("%v/%v_video.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		m.logger.Fatal(err)
	}
	m.fileVideo = fileVideo
}
func (m *MpegDash) closeFiles() {
	if err := m.fileAudio.Close(); err != nil {
		m.logger.Fatal(err)
	}
	if err := m.fileVideo.Close(); err != nil {
		m.logger.Fatal(err)
	}
}
func (m *MpegDash) getBaseUrl() string {
	return fmt.Sprintf("%v://%v/", m.manifestUrl.Scheme, m.manifestUrl.Host)
}
func (m *MpegDash) getRelativePath() string {
	splitedPath := strings.Split(m.manifestUrl.Path, "/")
	cutedPath := strings.Join(splitedPath[:len(splitedPath)-1], "/")
	return fmt.Sprintf("%v%v/", strings.TrimRight(m.getBaseUrl(), "/"), cutedPath)
}
func (m *MpegDash) fetchManifest() {
	response, err := http.Get(m.manifestUrl.String())
	if err != nil {
		m.logger.Fatal(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			m.logger.Fatal(err)
		}
	}()

	manifest := new(mpd.MPD)
	sliceOf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		m.logger.Fatal(err)
	}

	if err = manifest.Decode(sliceOf); err != nil {
		m.logger.Fatal(err)
	}
	m.manifest = manifest

	startTime, err := time.Parse("2006-01-02T15:04:05", m.manifest.AvailabilityStartTime.String())
	if err != nil {
		m.logger.Fatal(err)
	}
	diffTime := time.Now().Unix() - startTime.Unix()

	videoMedia := ""
	audioMedia := ""
	for _, period := range m.manifest.Period {
		for _, adaptation := range period.AdaptationSets {
			timescale := *adaptation.SegmentTemplate.Timescale
			if adaptation.MimeType == "audio/mp4" {
				if audioMedia != "" {
					continue
				}
				if !m.initAudio {
					m.initFirstChunk(audio, *adaptation.SegmentTemplate.Initialization)
				}

				audioMedia = *adaptation.SegmentTemplate.Media
				audioDuration := *adaptation.SegmentTemplate.Duration
				audioStartNumber := *adaptation.SegmentTemplate.StartNumber
				audioTimescale := *adaptation.SegmentTemplate.Timescale

				m.media.SetByType(audio, audioDuration, audioTimescale, audioStartNumber, diffTime, audioMedia)
			}
			if adaptation.MimeType == "video/mp4" {
				for _, repres := range adaptation.Representations {
					if videoMedia != "" {
						continue
					}
					if !m.initVideo {
						m.initFirstChunk(video, *repres.SegmentTemplate.Initialization)
					}

					videoMedia = *repres.SegmentTemplate.Media
					duration := *repres.SegmentTemplate.Duration
					startNumber := *repres.SegmentTemplate.StartNumber

					m.media.SetByType(video, duration, timescale, startNumber, diffTime, videoMedia)
				}
			}
		}
	}
}
func (m *MpegDash) initFirstChunk(ft fileType, initUrl string) {
	iniUrl := m.getRelativePath() + initUrl
	file := m.fileVideo
	if ft == audio {
		file = m.fileAudio
	}
	if err := m.downloadFilePart(iniUrl, file); err != nil {
		m.logger.Fatal(err)
	}
	if ft == audio {
		m.initAudio = true
	} else {
		m.initVideo = true
	}
}
func (m *MpegDash) downloadFilePart(fullUrl string, f *os.File) error {
	resp, err := http.Get(fullUrl)
	if err != nil {
		m.logger.Fatal(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.Fatal(err)
		}
	}()

	if resp.StatusCode == 404 {
		fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
		return errors.New("not found chunk")
	}

	if _, err := io.Copy(f, io.TeeReader(resp.Body, m.counter)); err != nil {
		m.logger.Fatal(err)
	}
	return nil
}
func (m *MpegDash) collectorChunks(store []string, value string) []string {
	maxStoredPrevChunks := 10
	slicePrevChunks := 5

	if len(store) == maxStoredPrevChunks {
		store = append(store[len(store)-slicePrevChunks:], value)
	} else {
		store = append(store, value)
	}
	return store
}
