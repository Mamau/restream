package main

import (
	"github.com/mamau/restream/server"
)

func main() {
	server.InitLogger()
	server.Start()
}
