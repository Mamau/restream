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

func main() {
	//storage.InitLogger()
	//server.Start()
	var playList = "https://matchtv.ru/vdl/playlist/133529/adaptive/1605486224/bd73e46a3ad27f0f12af1051a498f66c/web.m3u8"
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

	//chunks = append(chunks, getChunkList(basePath, chunkListName)...)
	f, err := os.Create("./somefile.ts")
	if err != nil {
		log.Fatal("Download error: ", err)
	}
	defer f.Close()

	//prevChunks := make(map[string]string)
	//actualChunks := make(map[string]string)

	for {
		select {
		case chunks, ok := <-ch:
			if !ok {
				fmt.Println("ВСЕЕ")
				f.Close()
				return
			}
			//for k, v := range chunks {
			//	if _, ok := prevChunks[k]; ok {
			//		continue
			//	}
			//	prevChunks[k] = v
			//	actualChunks[k] = v
			//}
			fmt.Println("----WRITE----")
			downloadAndWrite(basePath, chunks, f)
			//actualChunks = make(map[string]string)
		}
	}
}

var prevChunks = make([]string, 0, 30)

func downloadAndWrite(basePath string, chunks []string, f *os.File) {
	for _, v := range chunks {
		if _, ok := Find(prevChunks, v); ok {
			fmt.Printf("%v <-----EXISTS\n", v)
			continue
		}
		fmt.Println(v)
		part, err := downloadFilePart(fmt.Sprintf("%v/%v", basePath, v))
		if err != nil {
			log.Fatal("download part error: ", err)
		}

		if _, err = f.Write(part); err != nil {
			log.Fatal("write part to output file: ", err)
		}
		if len(prevChunks) == 30 {
			fmt.Println("flush slice!")
			prevChunks = append(prevChunks[len(prevChunks)-20:], v)
		} else {
			prevChunks = append(prevChunks, v)
		}
		fmt.Println(prevChunks)
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

func downloadFilePart(url string) ([]byte, error) {
	result := make([]byte, 0)
	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if result, err = ioutil.ReadAll(resp.Body); err != nil {
		return result, err
	}

	return result, err
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
