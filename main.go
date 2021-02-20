package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/mamau/restream/server"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:   "https://86a6e750f009479ca2aba656eeeb035a@o195631.ingest.sentry.io/5634880",
		Debug: true,
	}); err != nil {
		log.Printf("sentry.Init: %s\n", err)
	}

	server.Start()
}
