package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
	"github.com/rskv-p/mini/pkg/x_log"
)

// Service defines the microservice interface.
type Service interface {
	Init() error
	Start() error
	AddEndpoint(string, Handler, ...EndpointOpt) error
	AddGroup(string, ...GroupOpt) Group
	Info() Info
	Stats() Stats
	Reset()
	Stop() error
	Stopped() bool
}

// Internal service state.
type service struct {
	Config
	m               sync.Mutex
	id              string
	endpoints       []*Endpoint
	verbSubs        map[string]*nats.Subscription
	started         time.Time
	nc              *nats.Conn
	natsHandlers    handlers
	stopped         bool
	initialized     bool
	asyncDispatcher asyncCallbacksHandler
}

// GroupOpt is a functional option for groups.
type GroupOpt func(*groupOpts)

type groupOpts struct {
	queueGroup string
	qgDisabled bool
}

// ServiceIdentity holds metadata for a service instance.
type ServiceIdentity struct {
	Name     string            `json:"name"`
	ID       string            `json:"id"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
}

func AddService(nc *nats.Conn, config Config) Service {
	if config.Metadata == nil {
		config.Metadata = map[string]string{}
	}

	svc := &service{
		Config:    config,
		nc:        nc,
		id:        nuid.Next(),
		verbSubs:  make(map[string]*nats.Subscription),
		endpoints: make([]*Endpoint, 0),
		asyncDispatcher: asyncCallbacksHandler{
			cbQueue: make(chan func(), 100),
		},
	}
	return svc
}

func (s *service) Init() error {
	if err := s.Config.valid(); err != nil {
		return err
	}
	s.initialized = true
	return nil
}

func (s *service) Start() error {
	if !s.initialized {
		if err := s.Init(); err != nil {
			return err
		}
	}

	go s.asyncDispatcher.run()
	s.wrapConnectionEventCallbacks()

	if s.Config.Endpoint != nil {
		h := s.Config.Endpoint.Handler
		for i := len(s.Config.Middleware) - 1; i >= 0; i-- {
			h = s.Config.Middleware[i](h)
		}
		opts := []EndpointOpt{WithEndpointSubject(s.Config.Endpoint.Subject)}
		if s.Config.Endpoint.Metadata != nil {
			opts = append(opts, WithEndpointMetadata(s.Config.Endpoint.Metadata))
		}
		if s.Config.Endpoint.QueueGroup != "" {
			opts = append(opts, WithEndpointQueueGroup(s.Config.Endpoint.QueueGroup))
		} else if s.Config.QueueGroup != "" {
			opts = append(opts, WithEndpointQueueGroup(s.Config.QueueGroup))
		}
		if err := s.AddEndpoint("default", h, opts...); err != nil {
			s.asyncDispatcher.close()
			if s.OnError != nil {
				s.OnError(s, err)
			}
			// Log error using global logger
			x_log.Error().Err(err).Msg("failed to add default endpoint")
			return err
		}
		// Log successful registration using global logger
		x_log.Info().Str("subject", s.Config.Endpoint.Subject).Msg("default endpoint registered")
	}

	pingResp := Ping{
		ServiceIdentity: s.serviceIdentity(),
		Type:            PingResponseType,
	}

	handleVerb := func(verb Verb, valuef func() any) func(req Request) {
		return func(req Request) {
			resp, _ := json.Marshal(valuef())
			if err := req.Respond(resp); err != nil {
				if err := req.Error("500", fmt.Sprintf("error handling %s: %s", verb, err), nil); err != nil && s.Config.ErrorHandler != nil {
					s.asyncDispatcher.push(func() {
						s.Config.ErrorHandler(s, &NATSError{req.Subject(), err.Error()})
					})
				}
				if s.OnError != nil {
					s.OnError(s, err)
				}
				// Log error using global logger
				x_log.Error().Str("verb", verb.String()).Err(err).Msg("error responding to verb")
			}
		}
	}

	for verb, source := range map[Verb]func() any{
		InfoVerb:   func() any { return s.Info() },
		PingVerb:   func() any { return pingResp },
		StatsVerb:  func() any { return s.Stats() },
		HealthVerb: func() any { return map[string]string{"status": "ok", "type": HealthResponseType} },
		DocsVerb:   func() any { return s.collectDocs() },
	} {
		if err := s.addVerbHandlers(s.nc, verb, handleVerb(verb, source)); err != nil {
			s.asyncDispatcher.close()
			if s.OnError != nil {
				s.OnError(s, err)
			}
			// Log error using global logger
			x_log.Error().Err(err).Msg("failed to register verb")
			return err
		}
	}

	s.started = time.Now().UTC()

	if s.OnStart != nil {
		s.OnStart(s)
	}

	// Log service start using global logger
	x_log.Info().Str("name", s.Config.Name).Str("version", s.Config.Version).Str("id", s.id).
		Int("endpoints", len(s.endpoints)).Msg("service started")

	return nil
}

// reqHandler processes a single request and updates stats.
func (s *service) reqHandler(endpoint *Endpoint, req *request) {
	start := time.Now()
	endpoint.Handler.Handle(req)
	dur := time.Since(start)

	s.m.Lock()
	defer s.m.Unlock()

	endpoint.stats.LastRequestTime = time.Now().UTC()
	endpoint.stats.NumRequests++
	endpoint.stats.ProcessingTime += dur
	endpoint.stats.AverageProcessingTime = endpoint.stats.ProcessingTime / time.Duration(endpoint.stats.NumRequests)

	if dur < endpoint.stats.MinProcessingTime || endpoint.stats.MinProcessingTime == 0 {
		endpoint.stats.MinProcessingTime = dur
	}
	if dur > endpoint.stats.MaxProcessingTime {
		endpoint.stats.MaxProcessingTime = dur
	}

	if req.respondError != nil {
		endpoint.stats.NumErrors++
		endpoint.stats.LastError = req.respondError.Error()
	}
}

// AddGroup creates a new endpoint group.
func (s *service) AddGroup(name string, opts ...GroupOpt) Group {
	var o groupOpts
	for _, opt := range opts {
		opt(&o)
	}
	qg, noQ := resolveQueueGroup(o.queueGroup, s.Config.QueueGroup, o.qgDisabled, s.Config.QueueGroupDisabled)
	// Log group creation using global logger
	x_log.Info().Str("group", name).Str("queue_group", qg).Bool("queue_disabled", noQ).Msg("created endpoint group")
	return &group{
		service:            s,
		prefix:             name,
		queueGroup:         qg,
		queueGroupDisabled: noQ,
	}
}

// addVerbHandlers registers handlers for system verbs.
func (s *service) addVerbHandlers(nc *nats.Conn, verb Verb, handler HandlerFunc) error {
	kinds := []struct {
		name string
		kind string
		id   string
	}{
		{"all", "", ""},
		{"kind", s.Config.Name, ""},
		{verb.String(), s.Config.Name, s.id},
	}
	for _, k := range kinds {
		if err := s.addInternalHandler(nc, verb, k.kind, k.id, k.name, handler); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) addInternalHandler(nc *nats.Conn, verb Verb, kind, id, name string, handler HandlerFunc) error {
	subj, err := ControlSubject(verb, kind, id)
	if err != nil {
		_ = s.Stop()
		if s.OnError != nil {
			s.OnError(s, err)
		}
		// Log error using global logger
		x_log.Error().Str("verb", verb.String()).Err(err).Msg("failed to generate control subject")
		return err
	}

	s.verbSubs[name], err = nc.Subscribe(subj, func(msg *nats.Msg) {
		handler(&request{
			msg: msg,
		})
	})
	if err != nil {
		_ = s.Stop()
		if s.OnError != nil {
			s.OnError(s, err)
		}
		// Log error using global logger
		x_log.Error().Str("subject", subj).Err(err).Msg("failed to subscribe")
		return err
	}

	// Log subscription using global logger
	x_log.Info().Str("subject", subj).Str("verb", verb.String()).Msg("subscribed to control subject")

	return nil
}

// Stop gracefully stops the service and drains subscriptions.
func (s *service) Stop() error {
	s.m.Lock()
	if s.stopped {
		s.m.Unlock()
		return nil
	}
	s.stopped = true
	s.m.Unlock()

	// Stop all endpoints
	for _, e := range s.endpoints {
		if err := e.stop(); err != nil {
			if s.OnError != nil {
				s.OnError(s, err)
			}
			// Log error using global logger
			x_log.Error().Str("name", e.Name).Str("subject", e.Subject).Err(err).Msg("failed to stop endpoint")
			return err
		}
	}

	// Drain verb subscriptions
	for key, sub := range s.verbSubs {
		if sub.IsValid() && s.nc != nil && !s.nc.IsClosed() {
			if err := sub.Drain(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
				if s.OnError != nil {
					s.OnError(s, err)
				}
				// Log error using global logger
				x_log.Error().Str("subject", sub.Subject).Err(err).Msg("failed to drain subscription")
				return fmt.Errorf("draining %q: %w", sub.Subject, err)
			}
		} else {
			// Log skipping drain using global logger
			x_log.Warn().Str("subject", sub.Subject).Msg("skipping drain, conn closed or subscription invalid")
		}
		delete(s.verbSubs, key)
	}

	// Unwrap any wrapped connection callbacks
	unwrapConnectionEventCallbacks(s.nc, s.natsHandlers)

	// Log shutdown using global logger
	x_log.Info().Str("service", s.Config.Name).Msg("service stopped")

	// Fire lifecycle hook
	if s.OnStop != nil {
		s.OnStop(s)
	}

	// Finalize async queue
	if s.DoneHandler != nil {
		s.asyncDispatcher.push(func() { s.DoneHandler(s) })
	}
	s.asyncDispatcher.close()

	return nil
}

// Stopped returns true if service is stopped.
func (s *service) Stopped() bool {
	s.m.Lock()
	defer s.m.Unlock()
	return s.stopped
}

// Info returns the current service metadata.
func (s *service) Info() Info {
	s.m.Lock()
	defer s.m.Unlock()

	endpoints := make([]EndpointInfo, 0, len(s.endpoints))
	for _, e := range s.endpoints {
		meta := map[string]string{}
		for k, v := range e.Metadata {
			meta[k] = v
		}

		if e.Disabled {
			meta["disabled"] = "true"
		}

		if e.Doc != nil {
			if doc := e.Doc(e); doc != nil {
				if docJSON, err := json.Marshal(doc); err == nil {
					meta["doc"] = string(docJSON)
				}
			}
		}

		endpoints = append(endpoints, EndpointInfo{
			Name:       e.Name,
			Subject:    e.Subject,
			QueueGroup: e.QueueGroup,
			Metadata:   meta,
		})
	}

	return Info{
		ServiceIdentity: s.serviceIdentity(),
		Type:            InfoResponseType,
		Description:     s.Config.Description,
		Endpoints:       endpoints,
	}
}

// serviceIdentity returns the identity of this service instance.
func (s *service) serviceIdentity() ServiceIdentity {
	return ServiceIdentity{
		Name:     s.Config.Name,
		ID:       s.id,
		Version:  s.Config.Version,
		Metadata: s.Config.Metadata,
	}
}

// WithGroupQueueGroup sets a queue group for a group.
func WithGroupQueueGroup(qg string) GroupOpt {
	return func(o *groupOpts) { o.queueGroup = qg }
}

// WithGroupQueueGroupDisabled disables queue group usage.
func WithGroupQueueGroupDisabled() GroupOpt {
	return func(o *groupOpts) { o.qgDisabled = true }
}

// matchSubscriptionSubject matches a NATS subject to an endpoint.
func (s *service) matchSubscriptionSubject(subj string) (*Endpoint, bool) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, sub := range s.verbSubs {
		if sub.Subject == subj {
			return nil, true
		}
	}
	for _, e := range s.endpoints {
		if matchEndpointSubject(e.Subject, subj) {
			return e, true
		}
	}
	return nil, false
}

// matchEndpointSubject performs wildcard subject matching.
func matchEndpointSubject(pattern, subj string) bool {
	pt := strings.Split(pattern, ".")
	st := strings.Split(subj, ".")

	if len(pt) > len(st) {
		return false
	}

	for i := range pt {
		if i == len(pt)-1 && pt[i] == ">" {
			return true
		}
		if pt[i] != st[i] && pt[i] != "*" {
			return false
		}
	}
	return true
}
