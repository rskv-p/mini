// file: mini/messaging.go
package service

import (
	"errors"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/constant"
	"github.com/rskv-p/mini/transport"
)

// Respond publishes a response and removes the context entry.
func (s *Service) Respond(msg codec.IMessage, subject string) error {
	data, err := codec.Marshal(msg)
	if err != nil {
		return err
	}
	if err := s.opts.Transport.Publish(subject, data); err != nil {
		return err
	}
	s.opts.Context.Delete(msg.GetContextID())
	return nil
}

// Pub publishes a message to a selected node with retry logic.
func (s *Service) Pub(service string, msg codec.IMessage) error {
	nodeID, err := s.opts.Selector.Select(service)
	if err != nil {
		return err
	}
	msg.SetType(constant.MessageTypePublish)

	data, err := codec.Marshal(msg)
	if err != nil {
		return err
	}

	retries, interval := s.retryConfig()
	return s.retrySend("Pub", retries, interval, func() error {
		return s.opts.Transport.Publish(nodeID, data)
	})
}

// Req sends a request to a selected node with retry logic.
func (s *Service) Req(service string, msg codec.IMessage, handler transport.ResponseHandler) error {
	nodeID, err := s.opts.Selector.Select(service)
	if err != nil {
		return err
	}
	msg.SetType(constant.MessageTypeRequest)

	data, err := codec.Marshal(msg)
	if err != nil {
		return err
	}

	retries, interval := s.retryConfig()
	return s.retrySend("Req", retries, interval, func() error {
		return s.opts.Transport.Request(nodeID, data, handler)
	})
}

// Broadcast sends the message to all nodes of a given service.
func (s *Service) Broadcast(service string, msg codec.IMessage) error {
	svcs, err := s.opts.Registry.GetService(service)
	if err != nil {
		return err
	}
	if len(svcs) == 0 {
		return errors.New("broadcast: service not found")
	}

	msg.SetType(constant.MessageTypePublish)
	successCount := 0
	var lastErr error

	for _, svc := range svcs {
		for _, node := range svc.Nodes {
			copyMsg := msg.Copy()
			data, err := codec.Marshal(copyMsg)
			if err != nil {
				s.logger.WithContext(msg.GetContextID()).Warn("broadcast marshal error for node %s: %v", node.ID, err)
				continue
			}
			if err := s.opts.Transport.Publish(node.ID, data); err != nil {
				lastErr = err
				s.logger.WithContext(msg.GetContextID()).Warn("broadcast publish error to node %s: %v", node.ID, err)
			} else {
				successCount++
			}
		}
	}

	s.logger.WithContext(msg.GetContextID()).Debug("broadcast completed: %d successful", successCount)
	return lastErr
}

// retrySend executes fn with retry/backoff.
func (s *Service) retrySend(label string, retries int, interval time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i <= retries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			s.logger.Warn("%s attempt %d failed: %v", label, i+1, err)
			time.Sleep(interval)
			interval *= 2
		}
	}
	return lastErr
}

func (s *Service) retryConfig() (int, time.Duration) {
	retries := s.opts.Retry.Count
	if retries <= 0 {
		retries = 3
	}
	interval := s.opts.Retry.Interval
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	return retries, interval
}
