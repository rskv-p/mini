package core

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSError is returned when a NATS subscription fails.
type NATSError struct {
	Subject     string
	Description string
}

func (e *NATSError) Error() string {
	return fmt.Sprintf("%q: %s", e.Subject, e.Description)
}

// ErrHandler handles NATS-related service errors.
type ErrHandler func(Service, *NATSError)

// DoneHandler is called when the service stops.
type DoneHandler func(Service)

// StatsHandler returns custom data for stats endpoint.
type StatsHandler func(*Endpoint) any

// handlers stores original NATS connection callbacks.
type handlers struct {
	closed   nats.ConnHandler
	asyncErr nats.ErrHandler
}

// wrapConnectionEventCallbacks sets custom connection/error handlers.
func (s *service) wrapConnectionEventCallbacks() {
	s.m.Lock()
	defer s.m.Unlock()

	s.natsHandlers.closed = s.nc.ClosedHandler()
	s.nc.SetClosedHandler(func(c *nats.Conn) {
		if s.Logger != nil {
			s.Logger.Infow("NATS connection closed")
		}
		s.Stop()
		if s.natsHandlers.closed != nil {
			s.natsHandlers.closed(c)
		}
	})

	s.natsHandlers.asyncErr = s.nc.ErrorHandler()
	s.nc.SetErrorHandler(func(c *nats.Conn, sub *nats.Subscription, err error) {
		if sub == nil {
			if s.Logger != nil {
				s.Logger.Errorw("async NATS error (no subscription)", "err", err)
			}
			if s.natsHandlers.asyncErr != nil {
				s.natsHandlers.asyncErr(c, sub, err)
			}
			return
		}

		endpoint, match := s.matchSubscriptionSubject(sub.Subject)
		if !match {
			if s.Logger != nil {
				s.Logger.Errorw("async NATS error (unmatched subject)",
					"subject", sub.Subject,
					"err", err,
				)
			}
			if s.natsHandlers.asyncErr != nil {
				s.natsHandlers.asyncErr(c, sub, err)
			}
			return
		}

		if s.Config.ErrorHandler != nil {
			s.Config.ErrorHandler(s, &NATSError{
				Subject:     sub.Subject,
				Description: err.Error(),
			})
		}

		s.m.Lock()
		if endpoint != nil {
			endpoint.stats.NumErrors++
			endpoint.stats.LastError = err.Error()
		}
		s.m.Unlock()

		if s.Logger != nil {
			s.Logger.Errorw("NATS async error in endpoint",
				"subject", sub.Subject,
				"err", err,
			)
		}

		if stopErr := s.Stop(); stopErr != nil {
			if s.Logger != nil {
				s.Logger.Errorw("error stopping service after async NATS error",
					"subject", sub.Subject,
					"err", stopErr,
				)
			}
			if s.natsHandlers.asyncErr != nil {
				s.natsHandlers.asyncErr(c, sub, errors.Join(err, stopErr))
			}
		} else if s.natsHandlers.asyncErr != nil {
			s.natsHandlers.asyncErr(c, sub, err)
		}
	})
}

// unwrapConnectionEventCallbacks restores original NATS handlers.
func unwrapConnectionEventCallbacks(nc *nats.Conn, h handlers) {
	if nc.IsClosed() {
		return
	}
	nc.SetClosedHandler(h.closed)
	nc.SetErrorHandler(h.asyncErr)
}
