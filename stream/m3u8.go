package stream

import (
	"bufio"
	"fmt"
	"github.com/mamau/restream/helpers"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Chunk struct {
	Name  string
	Order int
	Value []byte
}

type M3u8 struct {
	playlist      string
	basePath      string
	chunkListName string
	name          string
	chunksChan    chan Chunk
	prevChunks    []string
	file          *os.File
	chunkListChan chan []string
	deadline      int
}

func NewM3u8(n, p string) *M3u8 {
	m := &M3u8{
		deadline:      65,
		name:          n,
		playlist:      p,
		chunksChan:    make(chan Chunk),
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

func (m *M3u8) setFile() {
	file, err := os.OpenFile(fmt.Sprintf("./%v.ts", m.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
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
				fmt.Printf("stop execJob %v\n", m.name)
				return
			}
			m.downloadChunks(chunks)
		}
	}
}

func (m *M3u8) deadlineJob() {
	time.AfterFunc(time.Duration(m.deadline)*time.Second, func() {
		close(m.chunkListChan)
	})
}

func (m *M3u8) receiveChunksList() {
	m.chunkListChan <- m.fetchChunkList()
	for {
		select {
		case <-time.Tick(10 * time.Second):
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
	if chunkLists == nil {
		log.Fatal("Chunk list not found")
	}

	chunkListFile := chunkLists[len(chunkLists)-1]
	splitPath := strings.Split(chunkListFile, "/")
	m.basePath = strings.Join(splitPath[:len(splitPath)-1], "/")
	m.chunkListName = splitPath[len(splitPath)-1]
}

func (m *M3u8) fetchChunkList() []string {
	var chunks []string
	fullPath := fmt.Sprintf("%v/%v", m.basePath, m.chunkListName)
	fmt.Println("fetching Chunk list...")
	resp, err := http.Get(fullPath)
	if err != nil {
		log.Fatal("cant execJob Chunk file: ", err)
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
	fmt.Printf("execJob Chunk %v...\n", chunkName)
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

	m.chunksChan <- Chunk{Name: chunkName, Order: order, Value: result}
}
