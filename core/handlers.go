package core

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rskv-p/mini/pkg/x_log"
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
// wrapConnectionEventCallbacks sets custom connection/error handlers.
// wrapConnectionEventCallbacks sets custom connection/error handlers.
// wrapConnectionEventCallbacks sets custom connection/error handlers.
// wrapConnectionEventCallbacks sets custom connection/error handlers.
func (s *service) wrapConnectionEventCallbacks() {
	s.m.Lock()
	defer s.m.Unlock()

	// Store the original closed handler and set a custom one
	s.natsHandlers.closed = s.nc.ClosedHandler()
	s.nc.SetClosedHandler(func(c *nats.Conn) {
		// Log connection closure with global logger
		x_log.Info().Msg("NATS connection closed")
		s.Stop()

		// Call the original handler if available
		if s.natsHandlers.closed != nil {
			s.natsHandlers.closed(c)
		}
	})

	// Store the original error handler and set a custom one
	s.natsHandlers.asyncErr = s.nc.ErrorHandler()
	s.nc.SetErrorHandler(func(c *nats.Conn, sub *nats.Subscription, err error) {
		if sub == nil {
			// Log async error when no subscription is present
			x_log.Error().Err(err).Msg("async NATS error (no subscription)")
			if s.natsHandlers.asyncErr != nil {
				s.natsHandlers.asyncErr(c, sub, err)
			}
			return
		}

		// Ensure that `sub.Subject` is valid before proceeding
		if sub.Subject == "" {
			x_log.Error().Msg("async NATS error (empty subject)")
			return
		}

		// Match the subscription subject
		endpoint, match := s.matchSubscriptionSubject(sub.Subject)
		if !match {
			// Log async error for unmatched subject and trigger error callback
			x_log.Error().Str("subject", sub.Subject).Err(err).Msg("async NATS error (unmatched subject)")

			// Log the subject matching failure for debugging purposes
			x_log.Info().
				Str("expectedPattern", sub.Subject).
				Str("receivedSubject", sub.Subject).
				Msg("Subject matching failed")

			// Trigger error callback if the subject does not match
			if s.Config.ErrorHandler != nil {
				// Make sure we have a valid endpoint
				if endpoint != nil {
					s.Config.ErrorHandler(s, &NATSError{
						Subject:     sub.Subject,
						Description: "unmatched subject",
					})
				} else {
					x_log.Error().Msg("Failed to handle error callback due to nil endpoint")
				}
			}

			if s.natsHandlers.asyncErr != nil {
				s.natsHandlers.asyncErr(c, sub, err)
			}
			return
		}

		// Call the error handler from the config
		if s.Config.ErrorHandler != nil {
			// Added a check to ensure we don't pass a nil endpoint
			if endpoint != nil {
				s.Config.ErrorHandler(s, &NATSError{
					Subject:     sub.Subject,
					Description: err.Error(),
				})
			} else {
				x_log.Error().Msg("Endpoint is nil, skipping error handler")
			}
		}

		// Lock and update the endpoint stats if the endpoint is valid
		s.m.Lock()
		if endpoint != nil {
			endpoint.stats.NumErrors++
			endpoint.stats.LastError = err.Error()
		} else {
			x_log.Error().Msg("Endpoint is nil, unable to update stats")
		}
		s.m.Unlock()

		// Log the async error in the endpoint
		x_log.Error().Str("subject", sub.Subject).Err(err).Msg("NATS async error in endpoint")

		// Attempt to stop the service after the error
		if stopErr := s.Stop(); stopErr != nil {
			x_log.Error().Str("subject", sub.Subject).Err(stopErr).Msg("error stopping service after async NATS error")
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

// Error handler, которая проверяет, соответствует ли тема, и вызывает обработчик только для совпадающих тем.
func (s *service) matchAndHandleError(sub *nats.Subscription, err error) {
	// Проверка на то, что тема подписки совпадает с ожидаемым шаблоном
	_, match := s.matchSubscriptionSubject(sub.Subject)
	if !match {
		// Логируем ошибку для несоответствующей темы и не вызываем обработчик ошибок
		x_log.Info().Str("expectedPattern", s.Config.Endpoint.Subject).Str("receivedSubject", sub.Subject).Msg("Subject matching failed")
		return
	}

	// В случае совпадения вызываем обработчик ошибок
	if s.Config.ErrorHandler != nil {
		s.Config.ErrorHandler(s, &NATSError{
			Subject:     sub.Subject,
			Description: err.Error(),
		})
	}

	// Логируем ошибку для правильной темы
	x_log.Error().Str("subject", sub.Subject).Err(err).Msg("NATS async error in endpoint")

	// Пробуем остановить сервис после ошибки
	if stopErr := s.Stop(); stopErr != nil {
		x_log.Error().Str("subject", sub.Subject).Err(stopErr).Msg("error stopping service after async NATS error")
	}
}
