// Package observability provides centralized observability components including metrics,
// tracing, and monitoring for the entire service.
package observability

import (
	"fmt"

	"github.com/genvmoroz/lale/service/internal/observability/mongo"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics encapsulates all service observability metrics.
type Metrics struct {
	Mongo *mongo.Metrics
}

// Config holds configuration for all service metrics.
type Config struct {
	Mongo mongo.Config
}

// DefaultConfig returns the default configuration for all service metrics.
func DefaultConfig() Config {
	return Config{
		Mongo: mongo.DefaultConfig(),
	}
}

// NewMetrics creates a new observability metrics component with all sub-metrics.
func NewMetrics(cfg Config) *Metrics {
	return &Metrics{
		Mongo: mongo.New(cfg.Mongo),
	}
}

// Register registers all service metrics with the provided Prometheus registerer.
func (m *Metrics) Register(reg prometheus.Registerer) error {
	if err := m.Mongo.Register(reg); err != nil {
		return fmt.Errorf("register mongo metrics: %w", err)
	}

	// Future metrics registration here:
	// if err := m.Redis.Register(reg); err != nil {
	//     return fmt.Errorf("register redis metrics: %w", err)
	// }

	return nil
}

func (m *Metrics) Unregister(reg prometheus.Registerer) bool {
	result := m.Mongo.Unregister(reg)

	// Future metrics unregistration here:
	// result = m.Redis.Unregister(reg) && result

	return result
}
