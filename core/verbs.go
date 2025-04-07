package core

// Verb defines built-in operation types.
type Verb int64

const (
	PingVerb   Verb = iota // Ping operation type
	StatsVerb              // Stats operation type
	InfoVerb               // Info operation type
	HealthVerb             // Health operation type
	DocsVerb               // Docs operation type
)

const (
	ErrorHeader        = "Nats-Service-Error"               // Error header
	ErrorCodeHeader    = "Nats-Service-Error-Code"          // Error code header
	InfoResponseType   = "io.nats.micro.v1.info_response"   // Info response type
	PingResponseType   = "io.nats.micro.v1.ping_response"   // Ping response type
	HealthResponseType = "io.nats.micro.v1.health_response" // Health response type
	DocsResponseType   = "io.nats.micro.v1.docs_response"   // Docs response type
)

// String returns the string representation of a verb.
func (v Verb) String() string {
	switch v {
	case PingVerb:
		return "PING" // PING operation
	case StatsVerb:
		return "STATS" // STATS operation
	case InfoVerb:
		return "INFO" // INFO operation
	case HealthVerb:
		return "HEALTH" // HEALTH operation
	case DocsVerb:
		return "DOCS" // DOCS operation
	default:
		return ""
	}
}

// Ping represents the response format for PING requests.
type Ping struct {
	ServiceIdentity
	Type string `json:"type"` // Type of the PING response
}

// Info represents the response format for INFO requests.
type Info struct {
	ServiceIdentity
	Type        string         `json:"type"`        // Type of the INFO response
	Description string         `json:"description"` // Description of the service
	Endpoints   []EndpointInfo `json:"endpoints"`   // List of endpoints
}

// EndpointInfo describes a single endpoint.
type EndpointInfo struct {
	Name       string            `json:"name"`        // Endpoint name
	Subject    string            `json:"subject"`     // Endpoint subject
	QueueGroup string            `json:"queue_group"` // Endpoint queue group
	Metadata   map[string]string `json:"metadata"`    // Endpoint metadata
}

// collectDocs returns structured documentation for DOCS verb.
func (s *service) collectDocs() map[string]any {
	s.m.Lock()
	defer s.m.Unlock()

	docs := make(map[string]any)
	var documented, missing int

	// Collect docs for all endpoints
	for _, e := range s.endpoints {
		if e.Doc != nil {
			// If documentation function exists, collect it
			if doc := e.Doc(e); doc != nil {
				docs[e.Name] = doc
				documented++
			} else {
				missing++
				if s.Logger != nil {
					s.Logger.Warnw("doc function returned nil", "endpoint", e.Name)
				}
			}
		} else {
			missing++
			if s.Logger != nil {
				s.Logger.Warnw("endpoint is undocumented", "endpoint", e.Name)
			}
		}
	}

	// Log the collected documentation stats
	if s.Logger != nil {
		s.Logger.Infow("collected endpoint docs",
			"total", len(s.endpoints),
			"documented", documented,
			"missing", missing,
		)
	}

	// Return structured documentation
	return map[string]any{
		"type": DocsResponseType,
		"docs": docs,
	}
}
