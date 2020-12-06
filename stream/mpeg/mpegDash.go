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

type fileType string

const (
	video fileType = "video"
	audio          = "audio"
)

type chunkMpeg struct {
	Type     fileType
	Recorded bool
	Order    int
	Value    io.Reader
}
type chunkQueue struct {
	Next    int
	Storage []*chunkMpeg
}

func (c *chunkQueue) SetChunk(chunk *chunkMpeg) {
	c.Storage = append(c.Storage, chunk)
}
func (c *chunkQueue) GetNextCHunk() (*chunkMpeg, error) {
	//c.FlushRecorded()
	for _, v := range c.Storage {
		if c.Next == v.Order {
			c.Next++
			v.Recorded = true
			return v, nil
		}
	}
	return &chunkMpeg{}, errors.New("chunk not exists")
}
func (c *chunkQueue) FlushRecorded() {
	var recorded []int
	for i, v := range c.Storage {
		if v.Recorded {
			recorded = append(recorded, i)
		}
	}
	for _, v := range recorded {
		c.Storage[v] = c.Storage[len(c.Storage)-1]
		c.Storage[len(c.Storage)-1] = nil
		c.Storage = c.Storage[:len(c.Storage)-1]
	}
}

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
	chunkList   map[fileType]*chunkQueue
	chunksChan  chan *chunkMpeg
	manifest    *mpd.MPD
	manifestUrl *url.URL
	deadline    time.Duration
	fileAudio   *os.File
	fileVideo   *os.File
	logger      *storage.StreamLogger
}

func NewMpegDash(name, folder string, manifestUrl *url.URL) *MpegDash {
	queue := make(map[fileType]*chunkQueue)
	queue[video] = &chunkQueue{}
	queue[audio] = &chunkQueue{}

	m := &MpegDash{
		media:       &media{},
		name:        name,
		manifestUrl: manifestUrl,
		folder:      folder,
		processStop: make(chan bool),
		logger:      storage.NewStreamLogger(folder, name),
		chunksChan:  make(chan *chunkMpeg),
		chunkList:   queue,
	}
	m.deadlineJob()
	return m
}
func (m *MpegDash) Start() {
	m.setFiles()
	m.fetchManifest()

	for {
		select {
		//case chunkData := <-m.chunksChan:
		//	m.writeChunkToFile(chunkData)
		case <-m.processStop:
			return
		case <-time.Tick(1 * time.Second):
			if !m.stopped {
				//m.fetchManifest()
				currentMediaChunk := m.media.GetMedia()
				if _, ok := helpers.Find(m.chunksVideo, currentMediaChunk); ok {
					return
				}
				fmt.Println("Tick", m.videoOrder)
				m.chunksVideo = m.collectorChunks(m.chunksVideo, currentMediaChunk)

				urlVideo := m.getBaseUrl() + currentMediaChunk
				chunkData := &chunkMpeg{Type: video, Order: m.videoOrder}

				go m.downloadFilePart(urlVideo, chunkData)
				m.videoOrder++
			}
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
	for _, period := range m.manifest.Period {
		for _, adaptation := range period.AdaptationSets {
			timescale := *adaptation.SegmentTemplate.Timescale

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

					//m.startFetchingChunk(repres, timescale, diffTime)
					//videoMedia = "www"
				}
			}

		}
	}

	//go m.fetchAudio(manifest)
	//go m.fetchVideo(manifest)
}

//func (m *MpegDash) startFetchingChunk(repres mpd.Representation, timescale uint64, diffTime int64) {
//	med := *repres.SegmentTemplate.Media
//	duration := *repres.SegmentTemplate.Duration
//	startNumber := *repres.SegmentTemplate.StartNumber
//
//	res := strings.Split(med, "/")
//	videoMedia := strings.Join(res[3:], "/")
//
//	videoMedia = m.media.formula(duration, timescale, startNumber, diffTime, videoMedia)
//	//videoMedia = m.formula(duration, timescale, startNumber, diffTime, videoMedia)
//	if _, ok := helpers.Find(m.chunksVideo, videoMedia); ok {
//		return
//	}
//	m.chunksVideo = m.collectorChunks(m.chunksVideo, videoMedia)
//
//	//media.SetMedia(videoMedia)
//
//	urlVideo := m.getBaseUrl() + videoMedia
//	chunkData := &chunkMpeg{Type: video, Order: m.videoOrder}
//	go m.downloadFilePart(urlVideo, chunkData)
//	m.videoOrder++
//}
func (m *MpegDash) initFirstVideo(initUrl string) {
	iniUrl := m.getRelativePath() + initUrl
	chunkData := &chunkMpeg{Type: video, Order: m.videoOrder}
	m.downloadFilePart(iniUrl, chunkData)
	m.initVideo = true
	m.videoOrder++
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
						chunkData := &chunkMpeg{Type: video, Order: m.videoOrder}
						go m.downloadFilePart(iniUrl, chunkData)
						m.initVideo = true
						m.videoOrder++
					}

					med := *repres.SegmentTemplate.Media
					duration := *repres.SegmentTemplate.Duration
					startNumber := *repres.SegmentTemplate.StartNumber
					timescale := *adaptation.SegmentTemplate.Timescale

					res := strings.Split(med, "/")
					videoMedia = strings.Join(res[3:], "/")

					videoMedia = m.media.formula(duration, timescale, startNumber, diffTime, videoMedia)
					if _, ok := helpers.Find(m.chunksVideo, videoMedia); ok {
						return
					}
					m.chunksVideo = m.collectorChunks(m.chunksVideo, videoMedia)
				}
			}
		}
	}
	urlVideo := m.getBaseUrl() + videoMedia
	chunkData := &chunkMpeg{Type: video, Order: m.videoOrder}
	go m.downloadFilePart(urlVideo, chunkData)
	m.videoOrder++
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
					chunkData := &chunkMpeg{Type: audio, Order: m.audioOrder}
					go m.downloadFilePart(iniUrl, chunkData)
					m.initAudio = true
					m.audioOrder++
				}

				media := *adaptation.SegmentTemplate.Media
				duration := *adaptation.SegmentTemplate.Duration
				startNumber := *adaptation.SegmentTemplate.StartNumber
				timeScale := *adaptation.SegmentTemplate.Timescale

				urlMedia := strings.Split(media, "/")
				audioMedia = strings.Join(urlMedia[3:], "/")
				audioMedia = m.media.formula(duration, timeScale, startNumber, diffTime, audioMedia)

				if _, ok := helpers.Find(m.chunksAudio, audioMedia); ok {
					return
				}
				m.chunksAudio = m.collectorChunks(m.chunksAudio, audioMedia)
			}
		}
	}

	urlAudio := m.getBaseUrl() + audioMedia
	chunkData := &chunkMpeg{Type: audio, Order: m.audioOrder}
	go m.downloadFilePart(urlAudio, chunkData)
	m.audioOrder++
}
func (m *MpegDash) downloadFilePart(fullUrl string, chunkMpeg *chunkMpeg) {

	resp, err := http.Get(fullUrl)

	if err != nil {
		m.logger.FatalLogger.Fatalf("error while fetch url: %v, err: %v\n", fullUrl, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.FatalLogger.Fatalf("error while closing body response %v, err: %v\n", fullUrl, err)
		}
	}()

	chunkMpeg.Value = resp.Body
	queue := m.chunkList[chunkMpeg.Type]
	queue.SetChunk(chunkMpeg)

	if chunk, err := queue.GetNextCHunk(); err == nil {
		file := m.fileVideo
		if chunk.Type == audio {
			file = m.fileAudio
		}
		if _, err := io.Copy(file, chunk.Value); err != nil {
			m.logger.FatalLogger.Fatalf("write part to output file %v, erorr: %v\n", err, chunk.Type)
		}
		splittedUrl := strings.Split(fullUrl, "/")
		chunkName := splittedUrl[len(splittedUrl)-1]

		m.logger.InfoLogger.Printf("chunk name---------------------------- %v", chunkName)
		m.logger.InfoLogger.Printf("write chunk %v, order: %v\n", chunk.Type, chunk.Order)
	} else {
		m.logger.WarningLogger.Printf("NO CHUNK!?!?!?")
	}
}

//func (m *MpegDash) formula(duration, timescale, startNumber uint64, diffTime int64, media string) string {
//	formula := (int(diffTime) / (int(duration / timescale))) + int(startNumber)
//	num := fmt.Sprintf(media, formula)
//	res := strings.ReplaceAll(num, "$", "")
//	return strings.ReplaceAll(res, "Number", "")
//}
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
