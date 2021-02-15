package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/mamau/restream/server"
	"log"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:   "https://86a6e750f009479ca2aba656eeeb035a@o195631.ingest.sentry.io/5634880",
		Debug: true,
	})
	if err != nil {
		log.Printf("sentry.Init: %s\n", err)
	}

	server.Start()
}
