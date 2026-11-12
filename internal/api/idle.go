package api

import (
	"context"
	"net/http"
	"time"

	"myLlama/internal/runner"
)

// ServeHTTP accounts for queued and active work before idle shutdown can run.
// Read-only polling does not keep the model loaded.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.activityMu.Lock()
		h.active++
		h.activityMu.Unlock()
		defer func() {
			h.activityMu.Lock()
			h.active--
			h.lastActivity = time.Now()
			h.activityMu.Unlock()
		}()
	}
	h.serve.ServeHTTP(w, r)
}

// RunIdleMonitor stops only the model process, leaving the API available for an
// explicit start. Its caller owns cancellation and waits for it on shutdown.
func (h *Handler) RunIdleMonitor(ctx context.Context) {
	if h.config.IdleTimeout == 0 {
		return
	}
	// A fixed cadence also bounds overhead for sub-second configured timeouts.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			h.stopIfIdle(ctx, now)
		}
	}
}

func (h *Handler) stopIfIdle(ctx context.Context, now time.Time) {
	// Hold admission until Stop completes: a new chat must not race the idle
	// decision and have its process stopped underneath it.
	h.activityMu.Lock()
	defer h.activityMu.Unlock()
	if ctx.Err() != nil || h.active > 0 || now.Sub(h.lastActivity) < h.config.IdleTimeout || h.runner.Status().State != runner.Ready {
		return
	}
	stopCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if _, err := h.runner.Stop(stopCtx); err != nil {
		h.logger.Warn("idle runner shutdown failed", "error", err)
	} else {
		h.logger.Info("idle runner stopped", "idle_timeout", h.config.IdleTimeout.String())
	}
	h.lastActivity = time.Now()
}
