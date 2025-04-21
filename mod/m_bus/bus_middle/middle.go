package bus_middle

import (
	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
)

// Middleware represents a base middleware implementation.
type Middleware struct {
	handler func(*bus_req.Request) error
}

// Process executes the middleware logic.
func (m *Middleware) Process(req *bus_req.Request) error {
	if m.handler != nil {
		return m.handler(req)
	}
	return nil
}

// NewMiddleware creates a new middleware instance.
func NewMiddleware(handler func(*bus_req.Request) error) bus_type.IMiddleware {
	return &Middleware{handler: handler}
}
