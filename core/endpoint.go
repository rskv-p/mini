package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rskv-p/mini/pkg/x_log"
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
}

func (s *service) AddEndpoint(name string, handler Handler, opts ...EndpointOpt) error {
	var options endpointOpts
	options.handler = handler

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			x_log.Error().Str("name", name).Err(err).Msg("failed to apply endpoint option")
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

	x_log.Info().Str("name", name).
		Str("subject", subject).
		Str("queue_group", queueGroup).
		Bool("queue_disabled", noQueue).
		Bool("jetstream", options.useJetStream).
		Bool("disabled", options.disabled).
		Msg("adding endpoint")

	s.m.Lock()
	defer s.m.Unlock()

	for _, ep := range s.endpoints {
		if ep.Name == name {
			x_log.Error().Str("name", name).Msg("duplicate endpoint name")
			return fmt.Errorf("%w: duplicate endpoint name %q", ErrConfigValidation, name)
		}
		if ep.Subject == subject {
			x_log.Error().Str("subject", subject).Msg("duplicate endpoint subject")
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
		x_log.Info().Str("name", ep.Name).Str("subject", ep.Subject).Msg("endpoint is disabled, skipping subscription")
		return nil
	}

	if err := s.subscribeEndpoint(ep); err != nil {
		x_log.Error().Str("name", ep.Name).Str("subject", ep.Subject).Err(err).Msg("failed to subscribe endpoint")
		return err
	}

	s.endpoints = append(s.endpoints, ep)

	x_log.Info().Str("name", ep.Name).Str("subject", ep.Subject).Msg("endpoint subscribed")

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
) error {
	// validation with logging
	if !nameRegexp.MatchString(name) {
		x_log.Error().Str("name", name).Msg("invalid endpoint name")
		return fmt.Errorf("%w: invalid endpoint name", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(subject) {
		x_log.Error().Str("subject", subject).Msg("invalid endpoint subject")
		return fmt.Errorf("%w: invalid endpoint subject", ErrConfigValidation)
	}
	if queueGroup != "" && !subjectRegexp.MatchString(queueGroup) {
		x_log.Error().Str("queue_group", queueGroup).Msg("invalid queue group")
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
			msg: m,
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
		x_log.Error().Str("name", name).Str("subject", subject).Str("queue_group", queueGroup).Err(err).Msg("failed to subscribe to endpoint")
		return err
	}

	endpoint.subscription = sub
	endpoint.stats = EndpointStats{
		Name:       name,
		Subject:    subject,
		QueueGroup: queueGroup,
	}
	s.endpoints = append(s.endpoints, endpoint)

	x_log.Info().Str("name", name).Str("subject", subject).Str("queue_group", queueGroup).Msg("endpoint successfully subscribed")

	return nil
}

// stop unsubscribes and removes the endpoint from the service.
func (e *Endpoint) stop() error {
	s := e.service

	// Handle nil subscription gracefully
	if e.subscription == nil {
		x_log.Warn().Str("name", e.Name).Str("subject", e.Subject).Msg("stop called on endpoint with no active subscription")
		return nil
	}

	x_log.Info().Str("name", e.Name).Str("subject", e.Subject).Msg("stopping endpoint")

	// Skip drain if conn is closed or subscription invalid
	if e.subscription.IsValid() && s.nc != nil && !s.nc.IsClosed() {
		if err := e.subscription.Drain(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			x_log.Error().Str("name", e.Name).Str("subject", e.Subject).Err(err).Msg("failed to drain endpoint")
			return fmt.Errorf("draining %q: %w", e.Subject, err)
		}
	} else {
		x_log.Warn().Str("name", e.Name).Str("subject", e.Subject).Msg("skipping drain, conn closed or sub invalid")
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

	x_log.Info().Str("name", e.Name).Str("subject", e.Subject).Msg("endpoint successfully stopped")

	return nil
}

// WithEndpointSubject sets a custom subject.
func WithEndpointSubject(subject string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.subject = subject
		return nil
	}
}

// Additional methods for configuring endpoints remain unchanged...
// subscribeEndpoint subscribes the service to the endpoint.
func (s *service) subscribeEndpoint(ep *Endpoint) error {
	// Валидация
	if !nameRegexp.MatchString(ep.Name) {
		return fmt.Errorf("%w: invalid endpoint name", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(ep.Subject) {
		return fmt.Errorf("%w: invalid endpoint subject", ErrConfigValidation)
	}
	if ep.QueueGroup != "" && !subjectRegexp.MatchString(ep.QueueGroup) {
		return fmt.Errorf("%w: invalid queue group", ErrConfigValidation)
	}

	// Лог: начало подписки
	x_log.Info().Str("name", ep.Name).
		Str("subject", ep.Subject).
		Str("queue_group", ep.QueueGroup).
		Bool("jetstream", ep.useJetStream).
		Msg("subscribing to endpoint")

	cb := func(m *nats.Msg) {
		s.reqHandler(ep, &request{
			msg: m,
		})
	}

	var (
		sub *nats.Subscription
		err error
	)

	if ep.useJetStream {
		js, jserr := s.nc.JetStream()
		if jserr != nil {
			x_log.Error().Err(jserr).Msg("failed to get JetStream context")
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
		x_log.Error().Err(err).Str("name", ep.Name).
			Str("subject", ep.Subject).Str("queue_group", ep.QueueGroup).
			Msg("failed to subscribe to endpoint")
		return err
	}

	// Успешная подписка
	ep.subscription = sub
	ep.stats = EndpointStats{
		Name:       ep.Name,
		Subject:    ep.Subject,
		QueueGroup: ep.QueueGroup,
	}

	x_log.Info().Str("name", ep.Name).
		Str("subject", ep.Subject).
		Bool("jetstream", ep.useJetStream).
		Msg("subscription successful")

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

// WithJetStream enables JetStream for the endpoint.
func WithJetStream() EndpointOpt {
	return func(e *endpointOpts) error {
		e.useJetStream = true
		return nil
	}
}

// WithJetStreamConfig sets JetStream consumer config and enables JetStream.
// WithJetStreamConfig sets JetStream consumer config and enables JetStream.
func WithJetStreamConfig(cfg *nats.ConsumerConfig) EndpointOpt {
	return func(e *endpointOpts) error {
		e.useJetStream = true
		e.jsConsumer = cfg

		if cfg != nil {
			// Преобразуем длительность AckWait в миллисекунды (или секунды, в зависимости от предпочтений)
			ackWaitMillis := int(cfg.AckWait / time.Millisecond)

			x_log.Info().Str("durable", cfg.Durable).
				Int("ack_wait_ms", ackWaitMillis). // Логируем AckWait в миллисекундах
				Int("max_deliver", cfg.MaxDeliver).
				Str("ack_policy", ackPolicyToString(cfg.AckPolicy)).
				Str("deliver_policy", deliverPolicyToString(cfg.DeliverPolicy)).
				Msg("JetStream config applied")
		}

		return nil
	}
}

// ackPolicyToString converts AckPolicy to string.
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

// deliverPolicyToString converts DeliverPolicy to string.
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

// WithEndpointDisabled marks an endpoint as disabled via an option.
func WithEndpointDisabled() EndpointOpt {
	return func(e *endpointOpts) error {
		e.disabled = true
		// Log the action using global logger
		x_log.Debug().Msg("endpoint marked as disabled via option")
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

		// Wrap the handler with middlewares
		for i := len(mw) - 1; i >= 0; i-- {
			e.handler = mw[i](e.handler)
		}

		// Log the middleware application using global logger
		x_log.Info().
			Str("handler", originalType).
			Int("middleware_count", len(mw)).
			Msg("middleware applied")

		return nil
	}
}
