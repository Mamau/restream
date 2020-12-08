package mpeg

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"github.com/unki2aut/go-mpd"
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
		media:       &media{},
		name:        name,
		manifestUrl: manifestUrl,
		folder:      folder,
		processStop: make(chan bool),
		logger:      storage.NewStreamLogger(folder, name),
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
				m.fetchVideo()
				m.fetchAudio()
			}
		}
	}
}
func (m *MpegDash) fetchVideo() {
	currentMediaChunk := m.media.GetMedia()
	if _, ok := helpers.Find(m.chunksVideo, currentMediaChunk); !ok {
		urlVideo := m.getBaseUrl() + currentMediaChunk
		m.logger.InfoLogger.Println("Get chunk:", currentMediaChunk, "...")
		if err := m.downloadFilePart(urlVideo, m.fileVideo); err != nil {
			m.media.DecrementMedia()
		} else {
			m.chunksVideo = m.collectorChunks(m.chunksVideo, currentMediaChunk)
		}
	}
}

func (m *MpegDash) fetchAudio() {
	currentMediaChunk := m.media.GetAudioMedia()
	if _, ok := helpers.Find(m.chunksAudio, currentMediaChunk); !ok {
		urlAudio := m.getBaseUrl() + currentMediaChunk
		m.logger.InfoLogger.Println("Get chunk audio:", currentMediaChunk, "...")
		if err := m.downloadFilePart(urlAudio, m.fileAudio); err != nil {
			m.media.DecrementAudioMedia()
		} else {
			m.chunksVideo = m.collectorChunks(m.chunksVideo, currentMediaChunk)
		}
	}
}

func (m *MpegDash) Stop() {
	m.logger.InfoLogger.Println("stop mpeg dash")
	m.stopped = true
	time.AfterFunc(7*time.Second, m.closeFiles)
	//time.AfterFunc(10*time.Second, m.mergeAudioVideo)
	m.processStop <- true
}
func (m *MpegDash) SetDeadline(stopAt int64) {
	if stopAt <= time.Now().Unix() {
		m.logger.FatalLogger.Fatalf("deadline should be greater than now")
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
	m.logger.InfoLogger.Printf("start merge video %v and audio %v\n", m.fileVideo.Name(), m.fileAudio.Name())
	resultedFileName := fmt.Sprintf("%v/%v.mp4", m.folder, m.name)
	command := exec.Command("ffmpeg", "-i", m.fileVideo.Name(), "-i", m.fileAudio.Name(), "-c", "copy", resultedFileName)
	command.Stdout = m.logger.InfoLogger.Writer()
	command.Stderr = m.logger.WarningLogger.Writer()
	if err := command.Start(); err != nil {
		m.logger.ErrorLogger.Printf("cant merge files, error %v\n", err)
	}
}
func (m *MpegDash) setFiles() {
	if err := os.MkdirAll(m.folder, os.ModePerm); err != nil {
		m.logger.ErrorLogger.Printf("cant create folder, folder %v, error %v\n", m.folder, err)
	}

	fileAudio, err := os.OpenFile(fmt.Sprintf("%v/%v_audio.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		m.logger.FatalLogger.Fatalf("create audio file error: %v\n", err)
	}
	m.fileAudio = fileAudio

	fileVideo, err := os.OpenFile(fmt.Sprintf("%v/%v_video.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		m.logger.FatalLogger.Fatalf("create video file error: %v\n", err)
	}
	m.fileVideo = fileVideo
}
func (m *MpegDash) closeFiles() {
	if err := m.fileAudio.Close(); err != nil {
		m.logger.FatalLogger.Fatalf("cant close audio file, cause: %v", err)
	}
	if err := m.fileVideo.Close(); err != nil {
		m.logger.FatalLogger.Fatalf("cant close video file, cause: %v", err)
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
		m.logger.FatalLogger.Fatalf("fetch manifest file error: %v\n", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			m.logger.FatalLogger.Fatalf("error while closing manifest response: %v\n", err)
		}
	}()

	manifest := new(mpd.MPD)
	sliceOf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		m.logger.FatalLogger.Fatalf("cant read body %v\n", err)
	}

	if err = manifest.Decode(sliceOf); err != nil {
		m.logger.FatalLogger.Fatalf("error while decoding manifest file: %v\n", err)
	}
	m.manifest = manifest

	startTime, err := time.Parse("2006-01-02T15:04:05", m.manifest.AvailabilityStartTime.String())
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while parsing start time from manifest: %v\n", err)
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
					m.initFirstAudio(*adaptation.SegmentTemplate.Initialization)
				}

				audioMedia = *adaptation.SegmentTemplate.Media
				audioDuration := *adaptation.SegmentTemplate.Duration
				audioStartNumber := *adaptation.SegmentTemplate.StartNumber
				audioTimescale := *adaptation.SegmentTemplate.Timescale

				m.media.SetAudioMedia(audioDuration, audioTimescale, audioStartNumber, diffTime, audioMedia)
			}
			if adaptation.MimeType == "video/mp4" {
				for _, repres := range adaptation.Representations {
					if videoMedia != "" {
						continue
					}
					if !m.initVideo {
						m.initFirstVideo(*repres.SegmentTemplate.Initialization)
					}

					videoMedia = *repres.SegmentTemplate.Media
					duration := *repres.SegmentTemplate.Duration
					startNumber := *repres.SegmentTemplate.StartNumber

					m.media.SetMedia(duration, timescale, startNumber, diffTime, videoMedia)
				}
			}
		}
	}
}

func (m *MpegDash) initFirstVideo(initUrl string) {
	iniUrl := m.getRelativePath() + initUrl
	m.downloadFilePart(iniUrl, m.fileVideo)
	m.initVideo = true
}

func (m *MpegDash) initFirstAudio(initUrl string) {
	iniUrl := m.getRelativePath() + initUrl
	m.downloadFilePart(iniUrl, m.fileAudio)
	m.initAudio = true
}

func (m *MpegDash) downloadFilePart(fullUrl string, f *os.File) error {
	resp, err := http.Get(fullUrl)
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while fetch url: %v, err: %v\n", fullUrl, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.FatalLogger.Fatalf("error while closing body response %v, err: %v\n", fullUrl, err)
		}
	}()
	m.logger.InfoLogger.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
	if resp.StatusCode == 404 {
		return errors.New("not found chunk")
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	//if _, err := io.Copy(m.fileVideo, resp.Body); err != nil {
	//	m.logger.FatalLogger.Fatalf("write part to output file, erorr: %v\n", err)
	//}
	counts, _ := f.Write(bytes)
	m.logger.InfoLogger.Println("Writes: ------------ ", counts)
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
