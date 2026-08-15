package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// counterResponse mirrors the JSON returned by the rscounter service.
type counterResponse struct {
	Date       string `json:"date"`
	DayCount   int64  `json:"day_count"`
	TotalCount int64  `json:"total_count"`
}

// dayCount is one day's event total, from rscounter's history endpoint.
type dayCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// server holds the dependencies shared by the HTTP handlers.
type server struct {
	cfg    Config
	log    *slog.Logger
	client *http.Client
}

// newServer builds a server from config, wiring up the rscounter HTTP client.
func newServer(cfg Config, log *slog.Logger) *server {
	return &server{
		cfg:    cfg,
		log:    log,
		client: &http.Client{Timeout: cfg.RscounterTimeout},
	}
}

// routes registers all HTTP routes and returns the top-level handler.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/today", s.handleToday)
	mux.HandleFunc("GET /api/history", s.handleHistory)
	return s.withCORS(mux)
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleToday records a hit for today in rscounter and returns the date plus counts.
func (s *server) handleToday(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	counts, err := s.recordEvent(r.Context(), today)
	if err != nil {
		s.log.Error("rscounter call failed", "date", today, "err", err)
		http.Error(w, "counter service unavailable", http.StatusBadGateway)
		return
	}

	s.log.Info("served today",
		"date", counts.Date,
		"day_count", counts.DayCount,
		"total_count", counts.TotalCount,
	)
	writeJSON(w, http.StatusOK, counts)
}

// handleHistory returns the per-day event totals from rscounter.
func (s *server) handleHistory(w http.ResponseWriter, r *http.Request) {
	history, err := s.fetchHistory(r.Context())
	if err != nil {
		s.log.Error("rscounter history call failed", "err", err)
		http.Error(w, "counter service unavailable", http.StatusBadGateway)
		return
	}

	s.log.Info("served history", "days", len(history))
	writeJSON(w, http.StatusOK, history)
}

// fetchHistory gets the per-day totals from rscounter.
func (s *server) fetchHistory(ctx context.Context) ([]dayCount, error) {
	url := s.cfg.RscounterURL + "/history"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rscounter returned %d: %s", resp.StatusCode, msg)
	}

	var history []dayCount
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}
	return history, nil
}

// recordEvent posts an event to rscounter and returns the updated counts.
func (s *server) recordEvent(ctx context.Context, date string) (*counterResponse, error) {
	body, err := json.Marshal(map[string]string{"date": date})
	if err != nil {
		return nil, err
	}

	url := s.cfg.RscounterURL + "/events"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	s.log.Debug("calling rscounter", "url", url, "date", date)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rscounter returned %d: %s", resp.StatusCode, msg)
	}

	var counts counterResponse
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		return nil, err
	}
	return &counts, nil
}

// withCORS allows the React dev server (and any origin) to call this API.
func (s *server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
