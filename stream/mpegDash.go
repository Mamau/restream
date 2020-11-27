package stream

import (
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
	chunksAudio []string
	chunksVideo []string
	manifestUrl *url.URL
	deadline    time.Duration
	fileAudio   *os.File
	fileVideo   *os.File
	logger      *storage.StreamLogger
}

func NewMpegDash(name, folder string, manifestUrl *url.URL) *MpegDash {
	m := &MpegDash{
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
	for {
		select {
		case <-m.processStop:
			return
		case <-time.Tick(1 * time.Second):
			if !m.stopped {
				m.fetchManifest()
			}
		}
	}
}
func (m *MpegDash) Stop() {
	m.logger.InfoLogger.Println("stop mpeg dash")
	m.stopped = true
	m.closeFiles()
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
	m.logger.InfoLogger.Println("----------fetch manifest----------")
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
	err = manifest.Decode(sliceOf)
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while decoding manifest file: %v\n", err)
	}
	go m.fetchAudio(manifest)
	go m.fetchVideo(manifest)
}
func (m *MpegDash) fetchVideo(mpd *mpd.MPD) {
	videoMedia := ""
	startTime, err := time.Parse("2006-01-02T15:04:05", mpd.AvailabilityStartTime.String())
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while parsing start time from manifest: %v\n", err)
	}
	now := time.Now()
	diffTime := now.Unix() - startTime.Unix()

	for _, period := range mpd.Period {
		for _, adaptation := range period.AdaptationSets {
			if adaptation.MimeType == "video/mp4" {
				for _, repres := range adaptation.Representations {
					if videoMedia != "" {
						continue
					}
					if !m.initVideo {
						iniUrl := m.getRelativePath() + *repres.SegmentTemplate.Initialization
						m.logger.InfoLogger.Printf("%v<--------------INI VIDEO\n", iniUrl)
						m.fetchAndWriteToFile(iniUrl, m.fileVideo)
						m.initVideo = true
					}

					med := *repres.SegmentTemplate.Media
					duration := *repres.SegmentTemplate.Duration
					startNumber := *repres.SegmentTemplate.StartNumber
					timescale := *adaptation.SegmentTemplate.Timescale

					res := strings.Split(med, "/")
					videoMedia = strings.Join(res[3:], "/")

					videoMedia = m.formula(duration, timescale, startNumber, diffTime, videoMedia)
					if _, ok := helpers.Find(m.chunksVideo, videoMedia); ok {
						return
					}
					m.chunksVideo = m.collectorChunks(m.chunksVideo, videoMedia)
				}
			}
		}
	}
	urlVideo := m.getBaseUrl() + videoMedia
	m.logger.InfoLogger.Printf("%v<--------------WRITE VIDEO\n", urlVideo)
	m.fetchAndWriteToFile(urlVideo, m.fileVideo)
}
func (m *MpegDash) fetchAudio(mpd *mpd.MPD) {
	audioMedia := ""
	startTime, err := time.Parse("2006-01-02T15:04:05", mpd.AvailabilityStartTime.String())
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while parsing start time from manifest: %v\n", err)
	}
	now := time.Now()
	diffTime := now.Unix() - startTime.Unix()

	for _, period := range mpd.Period {
		for _, adaptation := range period.AdaptationSets {
			if adaptation.MimeType == "audio/mp4" {
				if audioMedia != "" {
					continue
				}
				if !m.initAudio {
					iniUrl := m.getRelativePath() + *adaptation.SegmentTemplate.Initialization
					m.logger.InfoLogger.Printf("%v<--------------INI AUDIO\n", iniUrl)
					m.fetchAndWriteToFile(iniUrl, m.fileAudio)
					m.initAudio = true
				}

				media := *adaptation.SegmentTemplate.Media
				duration := *adaptation.SegmentTemplate.Duration
				startNumber := *adaptation.SegmentTemplate.StartNumber
				timeScale := *adaptation.SegmentTemplate.Timescale

				urlMedia := strings.Split(media, "/")
				audioMedia = strings.Join(urlMedia[3:], "/")
				audioMedia = m.formula(duration, timeScale, startNumber, diffTime, audioMedia)

				if _, ok := helpers.Find(m.chunksAudio, audioMedia); ok {
					return
				}
				m.chunksAudio = m.collectorChunks(m.chunksAudio, audioMedia)
			}
		}
	}

	urlAudio := m.getBaseUrl() + audioMedia
	m.logger.InfoLogger.Printf("%v<--------------WRITE AUDIO\n", urlAudio)
	m.fetchAndWriteToFile(urlAudio, m.fileAudio)
}
func (m *MpegDash) fetchAndWriteToFile(url string, f *os.File) {
	resp, err := http.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.FatalLogger.Fatalf("error while closing body response %v, err: %v\n", url, err)
		}
	}()
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while fetch url: %v, err: %v\n", url, err)
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m.logger.FatalLogger.Fatalf("error while read response body url %v, err: %v\n", url, err)
	}

	if m.stopped {
		return
	}

	if _, err = f.Write(result); err != nil {
		m.logger.FatalLogger.Fatalf("error while write to file. err: %v\n", err)
	}
}
func (m *MpegDash) formula(duration, timescale, startNumber uint64, diffTime int64, media string) string {
	formula := (int(diffTime) / (int(duration / timescale))) + int(startNumber)
	num := fmt.Sprintf(media, formula)
	res := strings.ReplaceAll(num, "$", "")
	return strings.ReplaceAll(res, "Number", "")
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
