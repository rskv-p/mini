package core

import (
	"errors"
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"

	"github.com/nats-io/nats.go"
)

// Endpoint represents a registered service endpoint.
type Endpoint struct {
	EndpointConfig
	Name         string
	service      *service
	stats        EndpointStats
	subscription *nats.Subscription
	// JetStream
	useJetStream bool
	jsConsumer   *nats.ConsumerConfig
	Logger       x_log.Logger
}

// EndpointConfig holds configuration for an endpoint.
type EndpointConfig struct {
	Subject            string            // Subject to subscribe to
	Handler            Handler           // Handler for incoming requests
	Metadata           map[string]string `json:"metadata,omitempty"`
	QueueGroup         string            `json:"queue_group"`
	QueueGroupDisabled bool              `json:"queue_group_disabled"`
	Doc                DocFunc           `json:"-"`
	Disabled           bool              `json:"disabled"` // в EndpointConfig
}

// DocFunc allows generating documentation for the endpoint.
type DocFunc func(*Endpoint) *EndpointDoc

type EndpointDoc struct {
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Request     any      `json:"request,omitempty"`
	Response    any      `json:"response,omitempty"`
	Errors      []string `json:"errors,omitempty"`
}

// EndpointOpt allows customizing endpoint options.
type EndpointOpt func(*endpointOpts) error

type endpointOpts struct {
	subject      string
	metadata     map[string]string
	queueGroup   string
	qgDisabled   bool
	doc          DocFunc
	disabled     bool
	handler      Handler // Handler to wrap with middleware
	useJetStream bool
	jsConsumer   *nats.ConsumerConfig
	logger       x_log.Logger // ← добавили

}

func (s *service) AddEndpoint(name string, handler Handler, opts ...EndpointOpt) error {
	var options endpointOpts
	options.handler = handler
	options.logger = s.Logger // ← проброс логгера в опции

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			if s.Logger != nil {
				s.Logger.Errorw("failed to apply endpoint option", "name", name, "err", err)
			}
			return err
		}
	}

	subject := options.subject
	if subject == "" {
		subject = name
	}

	queueGroup, noQueue := resolveQueueGroup(
		options.queueGroup,
		s.Config.QueueGroup,
		options.qgDisabled,
		s.Config.QueueGroupDisabled,
	)

	if s.Logger != nil {
		s.Logger.Infow("adding endpoint",
			"name", name,
			"subject", subject,
			"queue_group", queueGroup,
			"queue_disabled", noQueue,
			"jetstream", options.useJetStream,
			"disabled", options.disabled,
		)
	}

	s.m.Lock()
	defer s.m.Unlock()

	for _, ep := range s.endpoints {
		if ep.Name == name {
			if s.Logger != nil {
				s.Logger.Errorw("duplicate endpoint name", "name", name)
			}
			return fmt.Errorf("%w: duplicate endpoint name %q", ErrConfigValidation, name)
		}
		if ep.Subject == subject {
			if s.Logger != nil {
				s.Logger.Errorw("duplicate endpoint subject", "subject", subject)
			}
			return fmt.Errorf("%w: duplicate endpoint subject %q", ErrConfigValidation, subject)
		}
	}

	finalHandler := options.handler
	if finalHandler == nil {
		finalHandler = handler
	}

	ep := &Endpoint{
		service:      s,
		Name:         name,
		useJetStream: options.useJetStream,
		jsConsumer:   options.jsConsumer,
		Logger:       s.Logger,
		EndpointConfig: EndpointConfig{
			Subject:            subject,
			Handler:            finalHandler,
			Metadata:           options.metadata,
			QueueGroup:         queueGroup,
			QueueGroupDisabled: noQueue,
			Doc:                options.doc,
			Disabled:           options.disabled,
		},
	}

	if ep.Disabled {
		s.endpoints = append(s.endpoints, ep)
		if s.Logger != nil {
			s.Logger.Infow("endpoint is disabled, skipping subscription", "name", ep.Name, "subject", ep.Subject)
		}
		return nil
	}

	if err := s.subscribeEndpoint(ep); err != nil {
		if s.Logger != nil {
			s.Logger.Errorw("failed to subscribe endpoint", "name", ep.Name, "subject", ep.Subject, "err", err)
		}
		return err
	}

	s.endpoints = append(s.endpoints, ep)

	if s.Logger != nil {
		s.Logger.Infow("endpoint subscribed", "name", ep.Name, "subject", ep.Subject)
	}

	return nil
}

// addEndpoint creates and subscribes the endpoint with logging support.
func addEndpoint(
	s *service,
	name string,
	subject string,
	handler Handler,
	metadata map[string]string,
	queueGroup string,
	noQueue bool,
	logger x_log.Logger,
) error {
	// validation with logging
	if !nameRegexp.MatchString(name) {
		if logger != nil {
			logger.Errorw("invalid endpoint name", "name", name)
		}
		return fmt.Errorf("%w: invalid endpoint name", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(subject) {
		if logger != nil {
			logger.Errorw("invalid endpoint subject", "subject", subject)
		}
		return fmt.Errorf("%w: invalid endpoint subject", ErrConfigValidation)
	}
	if queueGroup != "" && !subjectRegexp.MatchString(queueGroup) {
		if logger != nil {
			logger.Errorw("invalid queue group", "queue_group", queueGroup)
		}
		return fmt.Errorf("%w: invalid queue group", ErrConfigValidation)
	}

	endpoint := &Endpoint{
		service: s,
		EndpointConfig: EndpointConfig{
			Subject:            subject,
			Handler:            handler,
			Metadata:           metadata,
			QueueGroup:         queueGroup,
			QueueGroupDisabled: noQueue,
		},
		Name: name,
	}

	cb := func(m *nats.Msg) {
		s.reqHandler(endpoint, &request{
			msg:    m,
			logger: logger,
		})
	}

	var (
		sub *nats.Subscription
		err error
	)

	if noQueue {
		sub, err = s.nc.Subscribe(subject, cb)
	} else {
		sub, err = s.nc.QueueSubscribe(subject, queueGroup, cb)
	}

	if err != nil {
		if logger != nil {
			logger.Errorw("failed to subscribe to endpoint",
				"name", name,
				"subject", subject,
				"queue_group", queueGroup,
				"err", err,
			)
		}
		return err
	}

	endpoint.subscription = sub
	endpoint.stats = EndpointStats{
		Name:       name,
		Subject:    subject,
		QueueGroup: queueGroup,
	}
	s.endpoints = append(s.endpoints, endpoint)

	if logger != nil {
		logger.Infow("endpoint successfully subscribed",
			"name", name,
			"subject", subject,
			"queue_group", queueGroup,
		)
	}

	return nil
}

// stop unsubscribes and removes the endpoint from the service.
func (e *Endpoint) stop() error {
	s := e.service

	// Handle nil subscription gracefully
	if e.subscription == nil {
		if s.Logger != nil {
			s.Logger.Warnw("stop called on endpoint with no active subscription",
				"name", e.Name,
				"subject", e.Subject,
			)
		}
		return nil
	}

	if s.Logger != nil {
		s.Logger.Infow("stopping endpoint",
			"name", e.Name,
			"subject", e.Subject,
		)
	}

	// Skip drain if conn is closed or subscription invalid
	if e.subscription.IsValid() && s.nc != nil && !s.nc.IsClosed() {
		if err := e.subscription.Drain(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			if s.Logger != nil {
				s.Logger.Errorw("failed to drain endpoint",
					"name", e.Name,
					"subject", e.Subject,
					"err", err,
				)
			}
			return fmt.Errorf("draining %q: %w", e.Subject, err)
		}
	} else if s.Logger != nil {
		s.Logger.Warnw("skipping drain, conn closed or sub invalid",
			"name", e.Name,
			"subject", e.Subject,
		)
	}

	// Remove endpoint from service list
	s.m.Lock()
	for i, ep := range s.endpoints {
		if ep == e {
			s.endpoints = append(s.endpoints[:i], s.endpoints[i+1:]...)
			break
		}
	}
	s.m.Unlock()

	if s.Logger != nil {
		s.Logger.Infow("endpoint successfully stopped",
			"name", e.Name,
			"subject", e.Subject,
		)
	}

	return nil
}

// WithEndpointSubject sets a custom subject.
func WithEndpointSubject(subject string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.subject = subject
		return nil
	}
}

// WithEndpointMetadata sets custom metadata.
func WithEndpointMetadata(metadata map[string]string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.metadata = metadata
		return nil
	}
}

// WithEndpointQueueGroup sets a custom queue group.
func WithEndpointQueueGroup(queueGroup string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.queueGroup = queueGroup
		return nil
	}
}

// WithEndpointQueueGroupDisabled disables the queue group.
func WithEndpointQueueGroupDisabled() EndpointOpt {
	return func(e *endpointOpts) error {
		e.qgDisabled = true
		return nil
	}
}

// WithEndpointDoc attaches documentation to the endpoint.
func WithEndpointDoc(fn DocFunc) EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata["has_doc"] = "true"
		e.doc = fn
		return nil
	}
}

func WithEndpointDisabled() EndpointOpt {
	return func(e *endpointOpts) error {
		e.disabled = true
		if e.logger != nil {
			e.logger.Debugw("endpoint marked as disabled via option")
		}
		return nil
	}
}

// WithEndpointMiddleware wraps handler with given middlewares.
func WithEndpointMiddleware(mw []Middleware) EndpointOpt {
	return func(e *endpointOpts) error {
		if e.handler == nil {
			return fmt.Errorf("handler is not set (place WithEndpointMiddleware after core handler or WithEndpointAsync)")
		}

		originalType := fmt.Sprintf("%T", e.handler)

		for i := len(mw) - 1; i >= 0; i-- {
			e.handler = mw[i](e.handler)
		}

		if e.logger != nil {
			e.logger.Infow("middleware applied",
				"handler", originalType,
				"middleware_count", len(mw),
			)
		}

		return nil
	}
}

func (s *service) subscribeEndpoint(ep *Endpoint) error {
	// валидация
	if !nameRegexp.MatchString(ep.Name) {
		return fmt.Errorf("%w: invalid endpoint name", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(ep.Subject) {
		return fmt.Errorf("%w: invalid endpoint subject", ErrConfigValidation)
	}
	if ep.QueueGroup != "" && !subjectRegexp.MatchString(ep.QueueGroup) {
		return fmt.Errorf("%w: invalid queue group", ErrConfigValidation)
	}

	// лог: начало подписки
	if s.Logger != nil {
		s.Logger.Infow("subscribing to endpoint",
			"name", ep.Name,
			"subject", ep.Subject,
			"queue_group", ep.QueueGroup,
			"jetstream", ep.useJetStream,
		)
	}

	cb := func(m *nats.Msg) {
		s.reqHandler(ep, &request{
			msg:    m,
			logger: s.Logger,
		})
	}

	var (
		sub *nats.Subscription
		err error
	)

	if ep.useJetStream {
		js, jserr := s.nc.JetStream()
		if jserr != nil {
			if s.Logger != nil {
				s.Logger.Errorw("failed to get JetStream context", "err", jserr)
			}
			return jserr
		}
		consumer := ep.jsConsumer
		if consumer == nil {
			consumer = &nats.ConsumerConfig{
				Durable:       ep.Name,
				AckPolicy:     nats.AckExplicitPolicy,
				DeliverPolicy: nats.DeliverAllPolicy,
			}
		}
		sub, err = js.QueueSubscribe(ep.Subject, ep.QueueGroup, cb,
			nats.ManualAck(),
			nats.Durable(consumer.Durable),
			nats.AckWait(consumer.AckWait),
			nats.MaxDeliver(consumer.MaxDeliver),
		)
	} else {
		if ep.QueueGroupDisabled {
			sub, err = s.nc.Subscribe(ep.Subject, cb)
		} else {
			sub, err = s.nc.QueueSubscribe(ep.Subject, ep.QueueGroup, cb)
		}
	}

	if err != nil {
		if s.Logger != nil {
			s.Logger.Errorw("failed to subscribe to endpoint",
				"name", ep.Name,
				"subject", ep.Subject,
				"queue_group", ep.QueueGroup,
				"err", err,
			)
		}
		return err
	}

	// успешная подписка
	ep.subscription = sub
	ep.stats = EndpointStats{
		Name:       ep.Name,
		Subject:    ep.Subject,
		QueueGroup: ep.QueueGroup,
	}

	if s.Logger != nil {
		s.Logger.Infow("subscription successful",
			"name", ep.Name,
			"subject", ep.Subject,
			"jetstream", ep.useJetStream,
		)
	}

	return nil
}

// WithEndpointAsync wraps handler in a goroutine.
func WithEndpointAsync() EndpointOpt {
	return func(e *endpointOpts) error {
		if e.handler == nil {
			return fmt.Errorf("handler is not set")
		}
		h := e.handler
		e.handler = HandlerFunc(func(req Request) {
			go h.Handle(req)
		})
		return nil
	}
}

func WithJetStream() EndpointOpt {
	return func(e *endpointOpts) error {
		e.useJetStream = true
		return nil
	}
}

// WithJetStreamConfig sets JetStream consumer config and enables JetStream.
func WithJetStreamConfig(cfg *nats.ConsumerConfig) EndpointOpt {
	return func(e *endpointOpts) error {
		e.useJetStream = true
		e.jsConsumer = cfg

		if e.logger != nil && cfg != nil {
			e.logger.Infow("JetStream config applied",
				"durable", cfg.Durable,
				"ack_wait", cfg.AckWait,
				"max_deliver", cfg.MaxDeliver,
				"ack_policy", ackPolicyToString(cfg.AckPolicy),
				"deliver_policy", deliverPolicyToString(cfg.DeliverPolicy),
			)
		}

		return nil
	}
}

func ackPolicyToString(p nats.AckPolicy) string {
	switch p {
	case nats.AckNonePolicy:
		return "none"
	case nats.AckAllPolicy:
		return "all"
	case nats.AckExplicitPolicy:
		return "explicit"
	default:
		return "unknown"
	}
}

func deliverPolicyToString(p nats.DeliverPolicy) string {
	switch p {
	case nats.DeliverAllPolicy:
		return "all"
	case nats.DeliverLastPolicy:
		return "last"
	case nats.DeliverNewPolicy:
		return "new"
	case nats.DeliverByStartSequencePolicy:
		return "by_start_sequence"
	case nats.DeliverByStartTimePolicy:
		return "by_start_time"
	case nats.DeliverLastPerSubjectPolicy:
		return "last_per_subject"
	default:
		return "unknown"
	}
}
