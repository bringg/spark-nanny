package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// server is a simple http server that exposes health check endpoint
type server struct {
	srv *http.Server
}

// newServer creates a new server listening on the given address
func newServer(webListenAddress string) *server {
	mux := http.NewServeMux()

	s := &server{
		srv: &http.Server{
			Addr:    webListenAddress,
			Handler: mux,
		},
	}

	mux.HandleFunc("/healthz", s.healthz)
	mux.Handle("/metrics", promhttp.Handler())

	return s
}

// startAsync starts the server asynchronously
func (s *server) startAsync() {
	go func() {
		if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()
}

// stop gracefully shuts down the server
func (s *server) stop() {
	// we don't expect any long running connections,
	// so we can safely shutdown the server using background context
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown server")
	}
}

// healthz is a simple health check endpoint
func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) // nolint: errcheck
}
