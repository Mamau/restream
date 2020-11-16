package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type chunk struct {
	Order int
	Value []byte
}

var chunksChannel = make(chan chunk)
var prevChunks = make([]string, 0, 30)

func main() {
	//storage.InitLogger()
	//server.Start()
	var playList = "https://matchtv.ru/vdl/playlist/133529/adaptive/1605575479/a341547fb5bd5ac41d90ed79d004ee90/web.m3u8"
	//var chunks []string
	basePath, chunkListName := getChunkInfoFromPlayList(playList)
	ch := make(chan []string)

	go func() {
		//time.AfterFunc(65*time.Second, func() {
		//	fmt.Println("Stop it")
		//	close(ch)
		//})
		for {
			select {
			case <-time.Tick(10 * time.Second):
				//if _, ok := <- ch; ok {
				ch <- getChunkList(basePath, chunkListName)
				fmt.Println("Receive chunks")
				//}
				//return
			}
		}
	}()

	f, err := os.Create("./somefile.ts")
	if err != nil {
		log.Fatal("Download error: ", err)
	}
	defer f.Close()

	for {
		select {
		case chunks, ok := <-ch:
			if !ok {
				fmt.Println("ВСЕЕ")
				f.Close()
				return
			}
			downloadAndWrite(basePath, chunks, f)
		}
	}
}

func downloadAndWrite(basePath string, chunks []string, f *os.File) {
	chLen := 0
	fullLen := len(chunks)
	for i, v := range chunks {
		if _, ok := Find(prevChunks, v); ok {
			continue
		}
		chLen++
		log.Printf("%v index it", i)
		go downloadFilePart(fmt.Sprintf("%v/%v", basePath, v), i)

		if len(prevChunks) == 30 {
			fmt.Println("flush slice!")
			prevChunks = append(prevChunks[len(prevChunks)-20:], v)
		} else {
			prevChunks = append(prevChunks, v)
		}
	}

	chunkList := make([][]byte, fullLen, fullLen)
	received := 0

	for {
		select {
		case chunkData := <-chunksChannel:
			received++
			chunkList[chunkData.Order] = chunkData.Value
			if received == chLen {
				for _, v := range chunkList {
					if len(v) == 0 {
						continue
					}
					if _, err := f.Write(v); err != nil {
						log.Fatal("write part to output file: ", err)
					}
				}
				return
			}
		}
	}
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func getChunkInfoFromPlayList(p string) (basePath string, chunkListName string) {
	var chunkLists []string

	resp, err := http.Get(p)
	if err != nil {
		log.Fatal("get manifest file error: ", err)
	}

	if isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices; !isOk {
		log.Fatal("file manifest is not access")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal("cant close response body")
		}
	}()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "https") {
			chunkLists = append(chunkLists, line)
		}
	}
	if chunkLists == nil {
		log.Fatal("chunk list not found")
	}

	chunkListFile := chunkLists[len(chunkLists)-1]
	splitPath := strings.Split(chunkListFile, "/")
	basePath = strings.Join(splitPath[:len(splitPath)-1], "/")
	chunkListName = splitPath[len(splitPath)-1]

	return basePath, chunkListName
}

func getChunkList(basePath string, chunkListName string) []string {
	var chunks []string
	fullPath := fmt.Sprintf("%v/%v", basePath, chunkListName)
	resp, err := http.Get(fullPath)
	if err != nil {
		log.Fatal("cant download chunk file: ", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal("cant close response body")
		}
	}()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, ".ts") {
			chunks = append(chunks, line)
		}
	}
	return chunks
}

func downloadFilePart(url string, order int) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("cant get cunk %v, cause %v", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalf("cant close body %v", err)
		}
	}()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("cant parse cunk %v, cause %v", url, err)
	}

	chunksChannel <- chunk{Order: order, Value: result}
}
