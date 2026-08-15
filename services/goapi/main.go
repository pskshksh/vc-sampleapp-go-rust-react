// goapi is the API the React frontend talks to. It reports today's date and,
// for each request, records a hit in the rscounter service and returns the
// resulting per-day and all-time counts.
package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Render logs in the same format as rscounter (tracing_subscriber fmt).
	log := slog.New(newTracingHandler(os.Stdout, cfg.LogLevel, "goapi"))

	srv := newServer(cfg, log)

	log.Info("goapi starting", "addr", cfg.Addr, "rscounter_url", cfg.RscounterURL)
	if err := http.ListenAndServe(cfg.Addr, srv.routes()); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
