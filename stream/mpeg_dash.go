package stream

import (
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/unki2aut/go-mpd"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type MpegDash struct {
	name        string
	folder      string
	initAudio   bool
	initVideo   bool
	chunksAudio []string
	chunksVideo []string
	manifestUrl *url.URL
	deadline    time.Duration
	fileAudio   *os.File
	fileVideo   *os.File
}

func NewMpegDash(name, folder string, manifestUrl *url.URL) *MpegDash {
	m := &MpegDash{
		name:        name,
		manifestUrl: manifestUrl,
		folder:      folder,
	}
	m.setFiles()
	m.deadlineJob()
	return m
}

func (m *MpegDash) Start() {
	for {
		select {
		case <-time.Tick(1 * time.Second):
			m.fetchManifest()
		}
	}
}
func (m *MpegDash) Stop() {}
func (m *MpegDash) SetDeadline(stopAt int64) {
	if stopAt <= time.Now().Unix() {
		log.Fatal("deadline should be greater than now")
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
func (m *MpegDash) setFiles() {
	if err := os.MkdirAll(m.folder, os.ModePerm); err != nil {
		zap.L().Error("cant create folder",
			zap.String("folder", m.folder),
			zap.String("error", err.Error()),
		)
	}

	fileAudio, err := os.OpenFile(fmt.Sprintf("%v/%v_audio.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal("create audio file error: ", err)
	}
	m.fileAudio = fileAudio

	fileVideo, err := os.OpenFile(fmt.Sprintf("%v/%v_video.mp4", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal("create video file error: ", err)
	}
	m.fileVideo = fileVideo
}
func (m *MpegDash) closeFiles() {
	if err := m.fileAudio.Close(); err != nil {
		log.Fatalf("cant close audio file, cause: %v", err)
	}
	if err := m.fileVideo.Close(); err != nil {
		log.Fatalf("cant close video file, cause: %v", err)
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
	fmt.Println("----------fetch manifest----------")
	response, err := http.Get(m.manifestUrl.String())
	if err != nil {
		log.Fatal("fetch manifest file error: ", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Fatalf("error while closing manifest response: %v\n", err)
		}
	}()

	manifest := new(mpd.MPD)
	sliceOf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = manifest.Decode(sliceOf)
	if err != nil {
		log.Fatalf("error while decoding manifest file: %v\n", err)
	}
	go m.fetchAudio(manifest)
	go m.fetchVideo(manifest)
}
func (m *MpegDash) fetchVideo(mpd *mpd.MPD) {
	videoMedia := ""
	startTime, err := time.Parse("2006-01-02T15:04:05", mpd.AvailabilityStartTime.String())
	if err != nil {
		log.Fatalf("error while parsing start time from manifest: %v\n", err)
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
						iniUrl := m.getBaseUrl() + *repres.SegmentTemplate.Initialization
						fmt.Println(iniUrl, "<--------------INI VIDEO")
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
						fmt.Printf("chunk already leaded%v\n", videoMedia)
						return
					}
					m.chunksAudio = append(m.chunksVideo, videoMedia)
				}
			}
		}
	}
	urlVideo := m.getBaseUrl() + videoMedia
	fmt.Println(urlVideo, "<--------------WRITE VIDEO")
	m.fetchAndWriteToFile(urlVideo, m.fileVideo)
}
func (m *MpegDash) fetchAudio(mpd *mpd.MPD) {
	audioMedia := ""
	startTime, err := time.Parse("2006-01-02T15:04:05", mpd.AvailabilityStartTime.String())
	if err != nil {
		log.Fatalf("error while parsing start time from manifest: %v\n", err)
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
					iniUrl := m.getBaseUrl() + *adaptation.SegmentTemplate.Initialization
					fmt.Println(iniUrl, "<--------------INI AUDIO")
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
					fmt.Printf("chunk already leaded%v\n", audioMedia)
					return
				}
				m.chunksAudio = append(m.chunksAudio, audioMedia)
			}
		}
	}

	urlAudio := m.getBaseUrl() + audioMedia
	fmt.Println(urlAudio, "<--------------WRITE AUDIO")
	m.fetchAndWriteToFile(urlAudio, m.fileAudio)
}
func (m *MpegDash) fetchAndWriteToFile(url string, f *os.File) {
	resp, err := http.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalf("error while closing body response %v, err: %v\n", url, err)
		}
	}()
	if err != nil {
		log.Fatalf("error while fetch url: %v, err: %v\n", url, err)
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error while read response body url %v, err: %v\n", url, err)
	}

	if _, err = f.Write(result); err != nil {
		log.Fatalf("error while write to file. err: %v\n", err)
	}
}
func (m *MpegDash) formula(duration, timescale, startNumber uint64, diffTime int64, media string) string {
	formula := (int(diffTime) / (int(duration / timescale))) + int(startNumber)
	num := fmt.Sprintf(media, formula)
	res := strings.ReplaceAll(num, "$", "")
	return strings.ReplaceAll(res, "Number", "")
}
