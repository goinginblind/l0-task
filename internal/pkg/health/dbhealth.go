package health

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/goinginblind/l0-task/internal/pkg/logger"
)

// Pinger wraps the PingContext method.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// DBHealthChecker monitors the health of the database connection.
type DBHealthChecker struct {
	pinger        Pinger
	logger        logger.Logger
	isHealthy     atomic.Bool
	checkInterval time.Duration
	checkTimeout  time.Duration
}

// NewDBHealthChecker creates a new DBHealthChecker. It does not start the monitoring.
func NewDBHealthChecker(pinger Pinger, logger logger.Logger, checkInterval, checkTimeout time.Duration) *DBHealthChecker {
	return &DBHealthChecker{
		pinger:        pinger,
		logger:        logger,
		checkInterval: checkInterval,
		checkTimeout:  checkTimeout,
	}
}

// Start begins the continuous health monitoring in a background goroutine.
// It performs an initial check synchronously to set the initial state.
func (hc *DBHealthChecker) Start(ctx context.Context) {
	hc.logger.Infow("Starting DB health checker...")
	hc.checkHealth(ctx)

	go func() {
		ticker := time.NewTicker(hc.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.checkHealth(ctx)
			case <-ctx.Done():
				hc.logger.Infow("Stopping DB health checker.")
				return
			}
		}
	}()
}

// IsHealthy returns the current health status of the database
func (hc *DBHealthChecker) IsHealthy() bool {
	return hc.isHealthy.Load()
}

// MarkUnhealthy allows an external component (like a worker) to
// flag the connection as unhealthy without waiting for the next scheduled check.
func (hc *DBHealthChecker) MarkUnhealthy() {
	// Use CompareAndSwap to only log the first time it's marked unhealthy.
	if hc.isHealthy.CompareAndSwap(true, false) {
		hc.logger.Warnw("DB connection proactively marked as unhealthy by a worker.")
	}
}

// checkHealth performs a single health check.
func (hc *DBHealthChecker) checkHealth(ctx context.Context) {
	pingCtx, cancel := context.WithTimeout(ctx, hc.checkTimeout)
	defer cancel()

	err := hc.pinger.PingContext(pingCtx)
	wasHealthy := hc.isHealthy.Load()

	if err != nil {
		if wasHealthy {
			hc.logger.Errorw("Database connection lost", "error", err)
			hc.isHealthy.Store(false)
		}
		return
	}

	if !wasHealthy {
		hc.logger.Infow("Database connection restored")
		hc.isHealthy.Store(true)
	}
}
