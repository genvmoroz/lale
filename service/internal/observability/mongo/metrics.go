// Package mongo provides Prometheus metrics for MongoDB operations monitoring.
package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/event"
)

// Metrics encapsulates MongoDB operation metrics for Prometheus.
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec

	commandDurations map[int64]time.Time
	mu               sync.RWMutex
}

// Config holds configuration for MongoDB metrics.
type Config struct {
	// Namespace is the Prometheus namespace for metrics (default: "")
	Namespace string
	// Subsystem is the Prometheus subsystem for metrics (default: "mongo_client")
	Subsystem string
	// Buckets for request duration histogram (default: standard database buckets)
	Buckets []float64
}

// DefaultConfig returns the default configuration for MongoDB metrics.
func DefaultConfig() Config {
	return Config{
		Namespace: "",
		Subsystem: "mongo_client",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}
}

// New creates a new MongoDB metrics component.
func New(cfg Config) *Metrics {
	return &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "requests_total",
				Help:      "Total number of MongoDB operations",
			},
			[]string{"operation", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "request_duration_seconds",
				Help:      "MongoDB operation duration in seconds",
				Buckets:   cfg.Buckets,
			},
			[]string{"operation"},
		),
		commandDurations: make(map[int64]time.Time),
	}
}

// Register registers the MongoDB metrics with the provided Prometheus registerer.
func (m *Metrics) Register(reg prometheus.Registerer) error {
	if err := reg.Register(m.requestsTotal); err != nil {
		return fmt.Errorf("register requests total: %w", err)
	}
	if err := reg.Register(m.requestDuration); err != nil {
		return fmt.Errorf("register request duration: %w", err)
	}
	return nil
}

// Unregister removes the MongoDB metrics from the provided Prometheus registerer.
// This is useful for testing or cleanup scenarios.
func (m *Metrics) Unregister(reg prometheus.Registerer) bool {
	result := reg.Unregister(m.requestsTotal)
	result = reg.Unregister(m.requestDuration) && result
	return result
}

// CommandMonitor returns a MongoDB command monitor that records metrics.
func (m *Metrics) CommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			m.mu.Lock()
			m.commandDurations[evt.RequestID] = time.Now()
			m.mu.Unlock()
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			m.mu.Lock()
			startTime, ok := m.commandDurations[evt.RequestID]
			if ok {
				delete(m.commandDurations, evt.RequestID)
			}
			m.mu.Unlock()

			if ok {
				duration := time.Since(startTime).Seconds()
				m.requestDuration.WithLabelValues(evt.CommandName).Observe(duration)
				m.requestsTotal.WithLabelValues(evt.CommandName, "success").Inc()
			}
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			m.mu.Lock()
			startTime, ok := m.commandDurations[evt.RequestID]
			if ok {
				delete(m.commandDurations, evt.RequestID)
			}
			m.mu.Unlock()

			if ok {
				duration := time.Since(startTime).Seconds()
				m.requestDuration.WithLabelValues(evt.CommandName).Observe(duration)
				m.requestsTotal.WithLabelValues(evt.CommandName, "error").Inc()
			}
		},
	}
}
