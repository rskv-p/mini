// file: arc/service/metrics.go
package service

import (
	"strings"
)

// metricRecorder is an internal helper for scoped metric updates.
type metricRecorder struct {
	service *Service
	prefix  string
}

// ----------------------------------------------------
// Service metrics management
// ----------------------------------------------------

func (s *Service) IncMetric(name string) {
	s.AddMetric(name, 1)
}

func (s *Service) AddMetric(name string, delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name] += delta
}

func (s *Service) SetMetric(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name] = value
}

func (s *Service) ResetMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = make(map[string]int64)
}

func (s *Service) Metrics() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]int64, len(s.metrics))
	for k, v := range s.metrics {
		out[k] = v
	}
	return out
}

func (s *Service) ExportMetrics() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]float64, len(s.metrics))
	for k, v := range s.metrics {
		out[k] = float64(v)
	}
	return out
}

// ----------------------------------------------------
// Prefix-based metric recorder
// ----------------------------------------------------

// WithMetricPrefix returns a scoped metric recorder with a prefix.
func (s *Service) WithMetricPrefix(prefix string) *metricRecorder {
	return &metricRecorder{
		service: s,
		prefix:  strings.TrimSuffix(prefix, ".") + ".",
	}
}

func (m *metricRecorder) full(name string) string {
	return m.prefix + name
}

func (m *metricRecorder) Inc(name string) {
	m.service.IncMetric(m.full(name))
}

func (m *metricRecorder) Add(name string, delta int64) {
	m.service.AddMetric(m.full(name), delta)
}

func (m *metricRecorder) Set(name string, value int64) {
	m.service.SetMetric(m.full(name), value)
}
