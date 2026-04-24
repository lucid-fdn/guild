package evaluator

import (
	"context"
	"log/slog"
	"time"
)

type Worker struct {
	service  *Service
	interval time.Duration
	logger   *slog.Logger
}

func NewWorker(service *Service, interval time.Duration, logger *slog.Logger) *Worker {
	if interval <= 0 {
		interval = time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{service: service, interval: interval, logger: logger}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, ok, err := w.service.RunNext(); err != nil {
				w.logger.Warn("evaluation job failed", "error", err)
			} else if ok {
				w.logger.Debug("evaluation job processed")
			}
		}
	}
}
