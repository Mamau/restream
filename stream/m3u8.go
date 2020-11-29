package stream

import (
	"bufio"
	"fmt"
	"github.com/mamau/restream/helpers"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type chunkM3u8 struct {
	Name  string
	Order int
	Value []byte
}

type M3u8 struct {
	playlist      string
	basePath      string
	chunkListName string
	name          string
	folder        string
	chunksChan    chan chunkM3u8
	prevChunks    []string
	file          *os.File
	chunkListChan chan []string
	deadline      time.Duration
	stopped       bool
}

func NewM3u8(outputName, folder, playlist string) *M3u8 {
	m := &M3u8{
		name:          outputName,
		folder:        folder,
		playlist:      playlist,
		chunksChan:    make(chan chunkM3u8),
		prevChunks:    make([]string, 0, 30),
		chunkListChan: make(chan []string),
	}
	m.setFile()
	m.deadlineJob()
	return m
}

func (m *M3u8) Start() {
	m.fetchChunkInfoFromPlayList()
	go m.receiveChunksList()
	m.execJob()
}

func (m *M3u8) Stop() {
	m.stopped = true
	close(m.chunkListChan)
}

func (m *M3u8) SetDeadline(stopAt int64) {
	if stopAt <= time.Now().Unix() {
		log.Fatal("deadline should be greater than now")
	}
	m.deadline = time.Duration(stopAt-time.Now().Unix()) * time.Second
	m.deadlineJob()
}

func (m *M3u8) deadlineJob() {
	if m.deadline == 0 {
		return
	}
	time.AfterFunc(m.deadline, m.Stop)
}

func (m *M3u8) setFile() {
	if err := os.MkdirAll(m.folder, os.ModePerm); err != nil {
		zap.L().Error("cant create folder",
			zap.String("folder", m.folder),
			zap.String("error", err.Error()),
		)
	}

	file, err := os.OpenFile(fmt.Sprintf("%v/%v.ts", m.folder, m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal("create file error: ", err)
	}
	m.file = file
}

func (m *M3u8) closeFile() {
	if err := m.file.Close(); err != nil {
		log.Fatalf("cant close file, cause: %v", err)
	}
}

func (m *M3u8) execJob() {
	for {
		select {
		case chunks, ok := <-m.chunkListChan:
			if !ok {
				m.closeFile()
				fmt.Printf("stop exec job %v\n", m.name)
				return
			}
			m.downloadChunks(chunks)
		}
	}
}

func (m *M3u8) receiveChunksList() {
	m.chunkListChan <- m.fetchChunkList()
	for {
		select {
		case <-time.Tick(10 * time.Second):
			if m.stopped {
				return
			}
			m.chunkListChan <- m.fetchChunkList()
		}
	}
}

func (m *M3u8) collectPrevChunks(v string) {
	maxStoredPrevChunks := 30
	slicePrevChunks := 20

	if len(m.prevChunks) == maxStoredPrevChunks {
		m.prevChunks = append(m.prevChunks[len(m.prevChunks)-slicePrevChunks:], v)
	} else {
		m.prevChunks = append(m.prevChunks, v)
	}
}

func (m *M3u8) downloadChunks(chunks []string) {
	chunkLen := 0
	fullLen := len(chunks)
	for i, v := range chunks {
		if _, ok := helpers.Find(m.prevChunks, v); ok {
			continue
		}
		chunkLen++
		go m.downloadFilePart(v, i)
		m.collectPrevChunks(v)
	}
	m.writeChunks(fullLen, chunkLen)
}

func (m *M3u8) writeChunks(fullLen, chunkLen int) {
	chunkList := make([][]byte, fullLen, fullLen)
	received := 0

	for {
		select {
		case chunkData := <-m.chunksChan:
			fmt.Printf("receive chunk %v...\n", chunkData.Name)
			received++
			chunkList[chunkData.Order] = chunkData.Value
			if received == chunkLen {
				for _, v := range chunkList {
					if len(v) == 0 {
						continue
					}
					if _, err := m.file.Write(v); err != nil {
						log.Fatal("write part to output file: ", err)
					}
				}
				return
			}
		}
	}
}

func (m *M3u8) fetchChunkInfoFromPlayList() {
	var chunkLists []string
	chunkListFile := m.playlist

	fmt.Println("fetch manifest info...")
	resp, err := http.Get(m.playlist)
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
	if chunkLists != nil {
		chunkListFile = chunkLists[len(chunkLists)-1]
	}

	splitPath := strings.Split(chunkListFile, "/")
	m.basePath = strings.Join(splitPath[:len(splitPath)-1], "/")
	m.chunkListName = splitPath[len(splitPath)-1]
}

func (m *M3u8) fetchChunkList() []string {
	var chunks []string
	fullPath := fmt.Sprintf("%v/%v", m.basePath, m.chunkListName)
	fmt.Println("fetching chunkM3u8 list...")
	resp, err := http.Get(fullPath)
	if err != nil {
		log.Fatal("cant execJob chunkM3u8 file: ", err)
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

func (m *M3u8) downloadFilePart(chunkName string, order int) {
	url := fmt.Sprintf("%v/%v", m.basePath, chunkName)
	fmt.Printf("execJob chunkM3u8 %v...\n", chunkName)
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

	m.chunksChan <- chunkM3u8{Name: chunkName, Order: order, Value: result}
}
