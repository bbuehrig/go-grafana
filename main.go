package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Service struct {
	metrics appMetrics
	config  appConfig
}

func newService() *Service {
	service := &Service{}
	service.initMetrics()
	service.readConfig()
	return service
}

func main() {
	service := newService()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	service.recordMetrics(ctx)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("Shutting down...")
}
