package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	server *http.Server
}

func NewServer(port string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Info().Str("address", s.server.Addr).Msg("Starting metrics server")
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("Metrics server error")
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping metrics server")
	return s.server.Shutdown(ctx)
}
