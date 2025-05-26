// file: mini/metrics.go
package service

import (
	"strings"
)

// ----------------------------------------------------
// Service-level metrics management
// ----------------------------------------------------

// IncMetric increments a metric by 1.
func (s *Service) IncMetric(name string) {
	s.AddMetric(name, 1)
}

// AddMetric increases the metric by the specified delta.
func (s *Service) AddMetric(name string, delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name] += delta
}

// SetMetric sets the metric to a specific value.
func (s *Service) SetMetric(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name] = value
}

// ResetMetrics clears all recorded metrics.
func (s *Service) ResetMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = make(map[string]int64)
}

// Metrics returns a copy of all metrics as int64.
func (s *Service) Metrics() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string]int64, len(s.metrics))
	for k, v := range s.metrics {
		copy[k] = v
	}
	return copy
}

// ExportMetrics returns metrics converted to float64 for external systems.
func (s *Service) ExportMetrics() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string]float64, len(s.metrics))
	for k, v := range s.metrics {
		copy[k] = float64(v)
	}
	return copy
}

// ----------------------------------------------------
// Scoped metric recorder with prefix
// ----------------------------------------------------

// metricRecorder allows namespacing metrics with a prefix.
type metricRecorder struct {
	service *Service
	prefix  string
}

// WithMetricPrefix returns a metricRecorder scoped by prefix.
func (s *Service) WithMetricPrefix(prefix string) *metricRecorder {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	return &metricRecorder{
		service: s,
		prefix:  prefix,
	}
}

// full generates full metric name with prefix.
func (m *metricRecorder) full(name string) string {
	return m.prefix + name
}

// Inc increments the prefixed metric by 1.
func (m *metricRecorder) Inc(name string) {
	m.service.IncMetric(m.full(name))
}

// Add adds delta to the prefixed metric.
func (m *metricRecorder) Add(name string, delta int64) {
	m.service.AddMetric(m.full(name), delta)
}

// Set sets the prefixed metric to a value.
func (m *metricRecorder) Set(name string, value int64) {
	m.service.SetMetric(m.full(name), value)
}
