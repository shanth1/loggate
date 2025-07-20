package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shanth1/gotools/log"
)

type server struct {
	metricsAddr string
}

func New(metricsAddr string) *server {
	return &server{
		metricsAddr,
	}
}

func (s *server) Start(ctx context.Context) {
	logger := log.FromCtx(ctx)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(s.metricsAddr, nil); err != nil {
		logger.Fatal().Msg(fmt.Sprintf("failed to start metrics server: %v", err))
	}
}
