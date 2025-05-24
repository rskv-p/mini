// file: mini/handler.go
package service

import (
	dcont "context"
	"errors"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/constant"
	"github.com/rskv-p/mini/context"
	"github.com/rskv-p/mini/recover"
	"github.com/rskv-p/mini/router"
)

// ServerHandler processes incoming transport messages.
func (s *Service) ServerHandler(msg codec.IMessage) {
	defer recover.RecoverWithContext(s.name, "ServerHandler", msg)

	switch msg.GetType() {
	case constant.MessageTypeRequest:
		s.handleRequest(msg, msg.GetReplyTo())
	case constant.MessageTypeResponse:
		s.handleResponse(msg, nil)
	case constant.MessageTypePublish:
		s.handlePublish(msg)
	case constant.MessageTypeHealthCheck:
		s.handleHealthCheck(msg, msg.GetReplyTo())
	default:
		s.logger.WithContext(msg.GetContextID()).Warn("unknown message type: %s", msg.GetType())
	}
}

// handleHealthCheck replies to a health-check request.
func (s *Service) handleHealthCheck(msg codec.IMessage, replyTo string) {
	go func() {
		defer recover.RecoverWithContext(s.name, "HealthCheck", msg)

		code, result := healthCheck(s.config)
		resp := codec.NewJsonResponse(msg.GetContextID(), code)
		resp.SetBody(result)

		if data, err := codec.Marshal(resp); err == nil {
			_ = s.opts.Transport.Publish(replyTo, data)
		} else {
			s.logger.WithContext(msg.GetContextID()).Error("health marshal error: %v", err)
		}
	}()
}

// handleRequest routes and executes a service request.
func (s *Service) handleRequest(msg codec.IMessage, replyTo string) {
	defer recover.RecoverWithContext(s.name, "ServerRequest", msg)

	if replyTo != "" {
		msg.SetReplyTo(replyTo)
	}

	msg.SetContextID(s.opts.Context.Add(&context.Conversation{
		ID:      msg.GetContextID(),
		Request: msg.GetReplyTo(),
	}))

	handler, err := s.opts.Router.Dispatch(msg)
	if err != nil {
		s.IncMetric("errors_total")
		s.logger.WithContext(msg.GetContextID()).Warn("no handler for node: %s", msg.GetNode())

		resp := codec.NewJsonResponse(msg.GetContextID(), 404)
		resp.SetError(constant.ErrNotFound)
		_ = s.Respond(resp, msg.GetReplyTo())
		return
	}

	go func() {
		defer recover.RecoverWithContext(s.name, "RequestHandler", msg)

		ctx := s.messageContext(msg)
		handler = router.Wrap(handler, s.opts.HdlrWrappers)

		if herr := handler(ctx, msg, replyTo); herr != nil {
			s.IncMetric("errors_total")
			resp := codec.NewJsonResponse(msg.GetContextID(), herr.StatusCode)
			resp.SetError(errors.New(herr.Message))
			_ = s.Respond(resp, replyTo)
		} else {
			s.IncMetric("responses_success")
		}
	}()
}

// handlePublish routes a publish message without reply.
func (s *Service) handlePublish(msg codec.IMessage) {
	defer recover.RecoverWithContext(s.name, "PublishHandler", msg)

	handler, err := s.opts.Router.Dispatch(msg)
	if err != nil {
		s.logger.WithContext(msg.GetContextID()).Warn("dispatch error: %v", err)
		s.IncMetric("errors_total")
		return
	}

	go func() {
		defer recover.RecoverWithContext(s.name, "Publish.Inner", msg)

		ctx := s.messageContext(msg)
		handler = router.Wrap(handler, s.opts.HdlrWrappers)
		_ = handler(ctx, msg, "")
		s.IncMetric("publish_handled")
	}()
}

// handleResponse forwards the response to the original requester.
func (s *Service) handleResponse(msg codec.IMessage, raw []byte) {
	defer recover.RecoverWithContext(s.name, "ResponseHandler", msg)

	conv := s.opts.Context.Get(msg.GetContextID())
	if conv == nil {
		s.IncMetric("errors_total")
		panic("conversation not found")
	}

	replyTo := conv.Request
	if raw == nil {
		var err error
		raw, err = codec.Marshal(msg)
		if err != nil {
			s.logger.WithContext(msg.GetContextID()).Error("marshal response error: %v", err)
			s.IncMetric("errors_total")
			return
		}
	}

	if err := s.opts.Transport.Publish(replyTo, raw); err != nil {
		s.logger.WithContext(msg.GetContextID()).Error("failed to publish response: %v", err)
		s.IncMetric("errors_total")
		panic("failed to publish response: " + err.Error())
	}

	s.opts.Context.Delete(msg.GetContextID())
	s.IncMetric("responses_sent")
}

// messageContext builds a context.Context from msg.ContextID
func (s *Service) messageContext(msg codec.IMessage) dcont.Context {
	return dcont.WithValue(dcont.Background(), ContextIDKey, msg.GetContextID())
}
