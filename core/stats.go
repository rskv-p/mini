package core

import (
	"encoding/json"
	"time"
)

const StatsResponseType = "io.nats.micro.v1.stats_response" // MIME type for stats responses

// Stats contains runtime stats for all endpoints.
type Stats struct {
	ServiceIdentity
	Type      string           `json:"type"`      // Type of the stats response
	Started   time.Time        `json:"started"`   // Service start time
	Endpoints []*EndpointStats `json:"endpoints"` // List of endpoint stats
}

// EndpointStats holds runtime statistics for an endpoint.
type EndpointStats struct {
	Name                  string          `json:"name"`                          // Endpoint name
	Subject               string          `json:"subject"`                       // Endpoint subject
	QueueGroup            string          `json:"queue_group"`                   // Queue group for the endpoint
	NumRequests           int             `json:"num_requests"`                  // Number of requests
	NumErrors             int             `json:"num_errors"`                    // Number of errors
	LastError             string          `json:"last_error"`                    // Last error message
	ProcessingTime        time.Duration   `json:"processing_time"`               // Total processing time
	AverageProcessingTime time.Duration   `json:"average_processing_time"`       // Average processing time
	MinProcessingTime     time.Duration   `json:"min_processing_time,omitempty"` // Minimum processing time
	MaxProcessingTime     time.Duration   `json:"max_processing_time,omitempty"` // Maximum processing time
	LastRequestTime       time.Time       `json:"last_request_time,omitempty"`   // Last request time
	Data                  json.RawMessage `json:"data,omitempty"`                // Custom stats data
}

// Stats returns statistics for all registered endpoints.
func (s *service) Stats() Stats {
	s.m.Lock()
	defer s.m.Unlock()

	if s.Logger != nil {
		s.Logger.Infow("collecting service stats", // Log stats collection
			"endpoints", len(s.endpoints),
			"since", s.started.Format(time.RFC3339),
		)
	}

	stats := Stats{
		ServiceIdentity: s.serviceIdentity(),
		Type:            StatsResponseType,
		Started:         s.started,
		Endpoints:       make([]*EndpointStats, 0, len(s.endpoints)), // Initialize stats slice
	}

	for _, ep := range s.endpoints {
		endpointStats := &EndpointStats{
			Name:                  ep.stats.Name,
			Subject:               ep.stats.Subject,
			QueueGroup:            ep.stats.QueueGroup,
			NumRequests:           ep.stats.NumRequests,
			NumErrors:             ep.stats.NumErrors,
			LastError:             ep.stats.LastError,
			ProcessingTime:        ep.stats.ProcessingTime,
			AverageProcessingTime: ep.stats.AverageProcessingTime,
			MinProcessingTime:     ep.stats.MinProcessingTime,
			MaxProcessingTime:     ep.stats.MaxProcessingTime,
			LastRequestTime:       ep.stats.LastRequestTime,
		}

		// Serialize custom stats if handler exists
		if s.StatsHandler != nil {
			if data, err := json.Marshal(s.StatsHandler(ep)); err == nil {
				endpointStats.Data = data
			} else if s.Logger != nil {
				s.Logger.Errorw("failed to serialize custom stats", // Log error if custom stats serialization fails
					"endpoint", ep.Name,
					"err", err,
				)
			}
		}

		// Log warnings for endpoints with errors
		if ep.stats.NumErrors > 0 && s.Logger != nil {
			s.Logger.Warnw("endpoint has errors",
				"name", ep.Name,
				"errors", ep.stats.NumErrors,
				"last_error", ep.stats.LastError,
			)
		}

		stats.Endpoints = append(stats.Endpoints, endpointStats) // Add stats for endpoint
	}

	return stats // Return collected stats
}

// Reset clears all collected stats and resets start time.
func (s *service) Reset() {
	s.m.Lock()
	defer s.m.Unlock()

	// Reset stats for all endpoints
	for _, ep := range s.endpoints {
		ep.reset()
	}

	s.started = time.Now().UTC() // Reset start time

	// Log reset information
	if s.Logger != nil {
		s.Logger.Infow("stats reset",
			"timestamp", s.started.Format(time.RFC3339),
			"endpoints", len(s.endpoints),
		)
	}
}

// reset clears collected stats for a single endpoint.
func (e *Endpoint) reset() {
	e.stats = EndpointStats{
		Name:    e.stats.Name,
		Subject: e.stats.Subject, // Reset endpoint stats
	}
}
