package server

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mamau/restream/routes/api/v1"
	//apiMiddleware "github.com/mamau/restream/routes/middleware"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func newRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	//router.Use(apiMiddleware.IsNginxRestreamRunning)

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
		if err := srv.ListenAndServe(); err != nil {
			zap.L().Fatal("cant listen",
				zap.String("duration", err.Error()),
			)
		}
	}()
	fmt.Printf("Server starting at: http://localhost%v\n", srv.Addr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
