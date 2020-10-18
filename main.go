package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	server()
}

func server() {
	fmt.Fprintf(os.Stdout, "Server started...")
	http.HandleFunc("/", stream)
	http.HandleFunc("/live", live)
	http.HandleFunc("/favicon.ico", doNothing)
	err := http.ListenAndServe(":89", nil)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func doNothing(w http.ResponseWriter, r *http.Request){}

func send(comnd *exec.Cmd,channelCmd chan *exec.Cmd) {
	comnd.Stdout = os.Stdout
	comnd.Stderr = os.Stderr

	err := comnd.Start()
	if err != nil {
		log.Fatalf("Cant start rtmp stream %s\n", err)
	}

	channelCmd <- comnd
}
func live(w http.ResponseWriter, r *http.Request) {
	indexFile, _ := ioutil.ReadFile("./dist/index.html")
	_, err := w.Write(indexFile)
	if err != nil {
		log.Fatalf("Error wrtie response %v", err)
	}
}

func receive(channelCmd chan *exec.Cmd) {
	for {
		select {
		case <-time.After(3600 * time.Minute):
			command := <- channelCmd
			fmt.Println("Kill process with PID: ", command.Process.Pid)
			err := command.Process.Kill()
			if err != nil {
				log.Fatalf("Cant kill process %v, error: %v", command.Process.Pid, err)
			}
			errWait := command.Wait()
			if errWait != nil {
				log.Fatalf("Cant wait process %v, error: %v", command.Process.Pid, err)
			}
			fmt.Println("Closed channel after 3 sec")
			close(channelCmd)
			return
		}
	}
}

func stream(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	file := "https://matchtv.ru/vdl/playlist/133529/adaptive/1603043062/701fb3138953fc82e9b5b4cfd2fbb8b5/web.m3u8"
	rtmpAddress := "rtmp://0.0.0.0:1935/stream/mystream"

	if query.Get("file") == "" {
		fmt.Fprintf(os.Stdout, "File not passed, use default")
	} else {
		file = query.Get("file")
	}
	cmd := exec.Command("ffmpeg", "-i", file, "-c", "copy", "-f", "flv", rtmpAddress)

	channel := make(chan *exec.Cmd)
	go send(cmd, channel)
	go receive(channel)

	_, err := w.Write([]byte("<strong>Stream starting...</strong>"))
	if err != nil {
		log.Fatalf("Error wrtie response %v", err)
	}
	return
}