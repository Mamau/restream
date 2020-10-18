package server

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mamau/restream/server/api/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func serverStarted(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("<strong>Stream starting...</strong>"))
	if err != nil {
		log.Fatalf("Error wrtie response %v", err)
	}
}

func newRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/", serverStarted)

	router.Mount("/api/v1/", v1.NewRouter())

	return router
}

func Start() {
	handler := newRouter()
	srv := &http.Server{
		Addr:         ":89",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("Cant listen %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
