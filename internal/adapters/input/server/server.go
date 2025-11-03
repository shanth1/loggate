package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shanth1/gotools/log"
)

type server struct {
	infoAddr string
}

func New(infoAddr string) *server {
	return &server{
		infoAddr,
	}
}

func (s *server) Start(ctx context.Context) {
	logger := log.FromContext(ctx)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(s.infoAddr, nil); err != nil {
		logger.Fatal().Msg(fmt.Sprintf("failed to start metrics server: %v", err))
	}
}
