package main

import (
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/unki2aut/go-mpd"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var file, _ = os.OpenFile(fmt.Sprintf("%v.mp4", "1tv"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)

var file2, _ = os.OpenFile(fmt.Sprintf("%v.mp4", "1tv_audio"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)

func main() {
	//storage.InitLogger()
	//server.Start()
	for {
		select {
		case <-time.Tick(1 * time.Second):
			fmt.Println("Tick")
			DL()
		}
	}

	file.Close()
	file2.Close()
}

var basebase = "https://edge4.1internet.tv/dash-live11/streams/1tv/"
var base = "https://edge4.1internet.tv/"
var playList = "https://edge4.1internet.tv/dash-live11/streams/1tv/1tvdash.mpd?e=1606240600"
var initialization = ""

var initialization_audio = ""
var first = true
var first_audio = true
var receivedChunks []string
var receivedChunksAudio []string
var receivedChunksApp []string

func DL() {
	//ffmpeg -i https://s40403.cdn.ngenix.net/dash-live11/streams/1tv/1tvdash.mpd?e=1605982080 -codec copy outtttt.mp4
	resp, err := http.Get(playList)
	if err != nil {
		log.Fatal("get manifest file error: ", err)
	}
	//fmt.Println(base)
	//scanner := bufio.NewScanner(resp.Body)
	//for scanner.Scan() {
	//	line := scanner.Text()
	//	fmt.Println(line)
	//}
	defer resp.Body.Close()
	mpd := new(mpd.MPD)
	sliceOf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	mpd.Decode(sliceOf)

	go mdp(mpd)
	go mdp_audio(mpd)
	// merge audio and video
	// ffmpeg -i 1tv.mp4 -i 1tv_audio.mp4 final.mp4
}

func mdp_audio(mpd *mpd.MPD) {

	startTime := mpd.AvailabilityStartTime
	chunk_audio := ""

	test, err := time.Parse("2006-01-02T15:04:05", startTime.String())
	if err != nil {
		panic(err)
	}
	now := time.Now()
	diffTime := now.Unix() - test.Unix()

	for _, period := range mpd.Period {
		for _, adaptation := range period.AdaptationSets {
			if adaptation.MimeType == "audio/mp4" {
				if chunk_audio != "" {
					continue
				}
				if initialization_audio == "" {
					initialization_audio = *adaptation.SegmentTemplate.Initialization
				}
				med_audio := *adaptation.SegmentTemplate.Media
				duration_audio := *adaptation.SegmentTemplate.Duration
				startNumber_audio := *adaptation.SegmentTemplate.StartNumber
				timescale_audio := *adaptation.SegmentTemplate.Timescale

				res_audio := strings.Split(med_audio, "/")
				chunk_audio = strings.Join(res_audio[3:], "/")
				chunk_audio = formula(duration_audio, timescale_audio, startNumber_audio, diffTime, chunk_audio)

				if _, ok := helpers.Find(receivedChunksAudio, chunk_audio); ok {
					fmt.Printf("Chunk already leaded%v\n", chunk_audio)
					return
				}

				receivedChunksAudio = append(receivedChunksAudio, chunk_audio)
			}
		}
	}

	if first_audio {
		urlIni_audio := basebase + initialization_audio
		fmt.Println(urlIni_audio, "<--------------INI AUDIO")
		writeToFile(urlIni_audio, file2)

		first_audio = false
	}

	// AUDIO
	url_audio := base + chunk_audio
	fmt.Println(url_audio, "<--------------AUDIO")
	writeToFile(url_audio, file2)
}

func mdp(mpd *mpd.MPD) {
	startTime := mpd.AvailabilityStartTime
	chunk := ""

	test, err := time.Parse("2006-01-02T15:04:05", startTime.String())
	if err != nil {
		panic(err)
	}
	now := time.Now()
	diffTime := now.Unix() - test.Unix()

	for _, period := range mpd.Period {
		for _, adaptation := range period.AdaptationSets {

			//if adaptation.MimeType == "application/mp4" {
			//	if chunk != "" {
			//		continue
			//	}
			//	if initialization == "" {
			//		initialization = *adaptation.SegmentTemplate.Initialization
			//	}
			//	med := *adaptation.SegmentTemplate.Media
			//	duration := *adaptation.SegmentTemplate.Duration
			//	startNumber := *adaptation.SegmentTemplate.StartNumber
			//	timescale := *adaptation.SegmentTemplate.Timescale
			//
			//	res := strings.Split(med, "/")
			//	chunk = strings.Join(res[3:], "/")
			//
			//	chunk = formula(duration, timescale, startNumber, diffTime, chunk)
			//	if _, ok := helpers.Find(receivedChunks, chunk); ok {
			//		fmt.Printf("Chunk already leaded%v\n", chunk)
			//		return
			//	}
			//
			//	receivedChunks = append(receivedChunks, chunk)
			//}
			if adaptation.MimeType == "video/mp4" {
				for _, repres := range adaptation.Representations {
					if chunk != "" {
						continue
					}
					if initialization == "" {
						initialization = *repres.SegmentTemplate.Initialization
					}
					med := *repres.SegmentTemplate.Media
					duration := *repres.SegmentTemplate.Duration
					startNumber := *repres.SegmentTemplate.StartNumber
					timescale := *adaptation.SegmentTemplate.Timescale

					res := strings.Split(med, "/")
					chunk = strings.Join(res[3:], "/")

					chunk = formula(duration, timescale, startNumber, diffTime, chunk)
					if _, ok := helpers.Find(receivedChunks, chunk); ok {
						fmt.Printf("Chunk already leaded%v\n", chunk)
						return
					}

					receivedChunks = append(receivedChunks, chunk)
				}
			}
		}
	}

	if first {
		urlIni := basebase + initialization
		fmt.Println(urlIni, "<--------------INI VIDEO")
		writeToFile(urlIni, file)

		first = false
	}

	// VIDEO
	url := base + chunk
	fmt.Println(url, "<--------------VIDEO")
	writeToFile(url, file)
}

func writeToFile(url string, f *os.File) {
	resp, _ := http.Get(url)
	result, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(result2)
	f.Write(result)
	resp.Body.Close()
}

func formula(duration, timescale, startNumber uint64, diffTime int64, med string) string {
	part := duration / timescale
	part2 := (int(diffTime) / int(part)) + int(startNumber)

	num := fmt.Sprintf(med, part2)
	res := strings.ReplaceAll(num, "$", "")
	res = strings.ReplaceAll(res, "Number", "")

	return res
}
