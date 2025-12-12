package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/report"
)

type Server struct {
	addr string
	log  *slog.Logger

	mu   sync.RWMutex
	last *report.Report
}

func New(addr string, log *slog.Logger) *Server {
	return &Server{addr: addr, log: log}
}

func (s *Server) SetLast(r *report.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = r
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		last := s.last
		s.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"status": "ok"}
		if last != nil {
			resp["last_status"] = last.Status
			resp["last_started"] = last.Started
			resp["last_duration"] = last.Duration.String()
			resp["last_backend"] = last.Backend
			resp["last_reboot_required"] = last.RebootRequired
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	s.log.Info("health server listening", "addr", s.addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
