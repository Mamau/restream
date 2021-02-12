package server

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mamau/restream/routes/api/v1"
	"github.com/mamau/restream/stream/scheduler"
	"github.com/mamau/restream/stream/selenium/channel"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"
)

func newRouter() http.Handler {
	router := chi.NewRouter()

	//router.Use(middleware.RealIP)
	//router.Use(middleware.Logger)
	//router.Use(middleware.Recoverer)
	//router.Use(middleware.Timeout(60 * time.Second))
	//router.Use(apiMiddleware.IsNginxRestreamRunning)

	router.HandleFunc("/debug/pprof/*", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Mount("/api/v1/", v1.NewRouter())

	return router
}

func Start() {
	handler := newRouter()
	srv := &http.Server{
		Addr:    ":8089",
		Handler: handler,
		//ReadTimeout:  5 * time.Second,
		//WriteTimeout: 10 * time.Second,
		//IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("cant listen")
		}
	}()

	scheduleChannels()
	fmt.Printf("Server starting at: http://localhost%v\n", srv.Addr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func scheduleChannels() {
	for _, v := range []channel.Channel{channel.TNT, channel.FIRST} {
		fmt.Printf("schedule channel: %s \n", v)
		scheduler.CreateScheduledChannel(v)
	}
}
