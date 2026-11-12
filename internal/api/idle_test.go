package api

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"myLlama/internal/runner"
)

type idleRunner struct {
	mockRunner
	stops int
}

func (r *idleRunner) Stop(context.Context) (runner.Status, error) {
	r.stops++
	r.status = runner.Status{State: runner.Stopped}
	return r.status, nil
}
func TestIdleShutdownProtectsActiveWorkAndIgnoresPolling(t *testing.T) {
	r := &idleRunner{}
	cfg := DefaultConfig()
	cfg.IdleTimeout = time.Minute
	h := newHandler(r, slog.Default(), cfg)
	now := time.Now()
	h.lastActivity = now.Add(-2 * time.Minute)
	h.active = 1 // Includes queued work and streams, not just active generation.
	h.stopIfIdle(t.Context(), now)
	if r.stops != 0 {
		t.Fatal("stopped during active work")
	}
	h.active = 0
	h.lastActivity = now
	h.stopIfIdle(t.Context(), now)
	if r.stops != 0 {
		t.Fatal("stopped before timeout")
	}
	h.lastActivity = now.Add(-2 * time.Minute)
	h.stopIfIdle(t.Context(), now)
	h.stopIfIdle(t.Context(), now.Add(2*time.Minute))
	if r.stops != 1 {
		t.Fatalf("stops = %d", r.stops)
	}
}
