package main

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// tracingHandler is a slog.Handler that renders records in the same format as
// the Rust `tracing_subscriber` fmt layer used by rscounter, e.g.:
//
//	2026-08-15T13:48:02.497105Z  INFO goapi: served today day_count=3
//
// so goapi and rscounter produce visually identical logs.
type tracingHandler struct {
	mu     *sync.Mutex
	w      io.Writer
	level  slog.Leveler
	target string
	attrs  []slog.Attr
}

func newTracingHandler(w io.Writer, level slog.Leveler, target string) *tracingHandler {
	return &tracingHandler{mu: &sync.Mutex{}, w: w, level: level, target: target}
}

// ANSI styling matching tracing's default fmt output.
const (
	ansiReset  = "\033[0m"
	ansiDim    = "\033[2m"
	ansiItalic = "\033[3m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiPurple = "\033[35m"
)

// timestampFormat mirrors tracing's RFC 3339 output with microsecond precision.
const timestampFormat = "2006-01-02T15:04:05.000000Z07:00"

func (h *tracingHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// levelStyle returns the color and 5-wide, right-aligned label tracing uses.
func levelStyle(l slog.Level) (color, label string) {
	switch {
	case l >= slog.LevelError:
		return ansiRed, "ERROR"
	case l >= slog.LevelWarn:
		return ansiYellow, " WARN"
	case l >= slog.LevelInfo:
		return ansiGreen, " INFO"
	case l >= slog.LevelDebug:
		return ansiBlue, "DEBUG"
	default:
		return ansiPurple, "TRACE"
	}
}

func (h *tracingHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder

	ts := r.Time
	if ts.IsZero() {
		ts = time.Now()
	}

	// <dim>timestamp<reset>
	b.WriteString(ansiDim)
	b.WriteString(ts.UTC().Format(timestampFormat))
	b.WriteString(ansiReset)
	b.WriteByte(' ')

	// <color>LEVEL<reset>
	color, label := levelStyle(r.Level)
	b.WriteString(color)
	b.WriteString(label)
	b.WriteString(ansiReset)
	b.WriteByte(' ')

	// <dim>target<reset><dim>:<reset>
	b.WriteString(ansiDim)
	b.WriteString(h.target)
	b.WriteString(ansiReset)
	b.WriteString(ansiDim)
	b.WriteByte(':')
	b.WriteString(ansiReset)
	b.WriteByte(' ')

	// message
	b.WriteString(r.Message)

	// fields: <italic>key<reset><dim>=<reset>value
	for _, a := range h.attrs {
		writeField(&b, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeField(&b, a)
		return true
	})

	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

func writeField(b *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	b.WriteString(ansiItalic)
	b.WriteString(a.Key)
	b.WriteString(ansiReset)
	b.WriteString(ansiDim)
	b.WriteByte('=')
	b.WriteString(ansiReset)
	b.WriteString(a.Value.Resolve().String())
}

func (h *tracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

// WithGroup is a no-op: goapi logs a flat set of fields, matching rscounter.
func (h *tracingHandler) WithGroup(_ string) slog.Handler {
	return h
}
