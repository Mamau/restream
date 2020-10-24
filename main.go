package main

import (
	"github.com/mamau/restream/server"
	"github.com/mamau/restream/storage"
)

func main() {
	storage.InitLogger()
	server.Start()
}
