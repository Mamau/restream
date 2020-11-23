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

var file, _ = os.OpenFile(fmt.Sprintf("%v.mp4", "plssss"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)

//var file2, _ = os.OpenFile(fmt.Sprintf("%v.mp4", "plssss_audio"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)

func main() {
	//storage.InitLogger()
	//server.Start()
	for {
		select {
		case <-time.Tick(2 * time.Second):
			fmt.Println("Tick")
			DL()
		}
	}
	file.Close()
	//file2.Close()
}

var basebase = "https://cdn10.1internet.tv/dash-live11/streams/1tv/"
var initialization = ""

var initialization_audio = ""
var first = true
var receivedChunks []string

func DL() {
	//ffmpeg -i https://s40403.cdn.ngenix.net/dash-live11/streams/1tv/1tvdash.mpd?e=1605982080 -codec copy outtttt.mp4
	base := "https://cdn10.1internet.tv/"
	playList := "https://cdn10.1internet.tv/dash-live11/streams/1tv/1tvdash.mpd?e=1606161014"

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

	mdp(base, resp)
}
func mdp(base string, resp *http.Response) {
	defer resp.Body.Close()
	mpd := new(mpd.MPD)
	sliceOf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	mpd.Decode(sliceOf)
	startTime := mpd.AvailabilityStartTime
	chunk := ""
	chunk_audio := ""

	test, err := time.Parse("2006-01-02T15:04:05", startTime.String())
	if err != nil {
		panic(err)
	}
	now := time.Now()
	diffTime := now.Unix() - test.Unix()

	for _, period := range mpd.Period {
		for _, adaptation := range period.AdaptationSets {

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
			if adaptation.MimeType == "audio/mp4" {
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
			}
		}
	}

	if first {
		urlIni_audio := basebase + initialization_audio
		fmt.Println(urlIni_audio, "<--------------INI AUDIO")
		resp2, _ := http.Get(urlIni_audio)
		result2, _ := ioutil.ReadAll(resp2.Body)
		//fmt.Println(result2)
		file.Write(result2)
		resp2.Body.Close()

		urlIni := basebase + initialization
		fmt.Println(urlIni, "<--------------INI VIDEO")
		resp1, _ := http.Get(urlIni)
		result, _ := ioutil.ReadAll(resp1.Body)
		//fmt.Println(result)
		file.Write(result)
		resp1.Body.Close()

		first = false
	}

	// AUDIO
	url_audio := base + chunk_audio
	fmt.Println(url_audio, "<--------------AUDIO")
	resp3, err3 := http.Get(url_audio)
	if err3 != nil {
		log.Fatalf("cant get url %v", err)
	}

	result3, _ := ioutil.ReadAll(resp3.Body)
	//fmt.Println(result3)
	file.Write(result3)
	resp3.Body.Close()

	// VIDEO
	url := base + chunk
	fmt.Println(url, "<--------------VIDEO")
	resp2, err := http.Get(url)
	if err != nil {
		log.Fatalf("cant get url %v", err)
	}

	result2, _ := ioutil.ReadAll(resp2.Body)
	//fmt.Println(result2)
	file.Write(result2)
	resp2.Body.Close()

}

func formula(duration, timescale, startNumber uint64, diffTime int64, med string) string {
	part := duration / timescale
	part2 := (int(diffTime) / int(part)) + int(startNumber)

	num := fmt.Sprintf(med, part2)
	res := strings.ReplaceAll(num, "$", "")
	res = strings.ReplaceAll(res, "Number", "")

	return res
}
