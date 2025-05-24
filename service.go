// file: mini/service.go
package service

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/config"
	"github.com/rskv-p/mini/logger"
	"github.com/rskv-p/mini/registry"
	"github.com/rskv-p/mini/router"
	"github.com/rskv-p/mini/selector"
	"github.com/rskv-p/mini/transport"

	"github.com/google/uuid"
)

var _ IClient = (*Service)(nil)
var _ IServer = (*Service)(nil)

// IService combines client and server interfaces.
type IService interface {
	IClient
	IServer
}

// IClient exposes transport-related APIs.
type IClient interface {
	ID() string
	Name() string
	Version() string

	Options() Options
	Config() map[string]string
	Context() context.Context

	Pub(service string, msg codec.IMessage) error
	Req(service string, msg codec.IMessage, handler transport.ResponseHandler) error
	Respond(msg codec.IMessage, subject string) error

	SubscribeTopic(topic string, handler transport.MsgHandler) error
	SubscribePrefix(prefix string, handler transport.MsgHandler) error
}

// IServer exposes server lifecycle and routing APIs.
type IServer interface {
	Init(...Option) error
	Run() error
	Stop() error

	RegisterAction(name string, schema []InputSchemaField, fn ActionFunc)
	RegisterActions(...IAction)
	ServerHandler(codec.IMessage)
	ListActions() []string
	GetSchemas() map[string][]InputSchemaField
	GetOpenAPISchemas() map[string]any
	Broadcast(service string, msg codec.IMessage) error
	Use(mw Middleware)
}

// Service is the core implementation.
type Service struct {
	opts    Options
	config  *config.Config
	name    string
	version string
	id      string
	logger  logger.ILogger

	actions map[string]actionInfo
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	metrics     map[string]int64
	mu          sync.RWMutex
	middlewares []Middleware
}

// NewService creates a new service instance.
func NewService(name, version string, extra ...Option) *Service {
	cfg := config.MustLoadFromEnv()
	ctx, cancel := context.WithCancel(context.Background())

	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	subject := strings.ReplaceAll(name, "-", ".") + "." + version + "." + id

	s := &Service{
		name:        name,
		version:     version,
		id:          id,
		config:      cfg,
		ctx:         ctx,
		cancel:      cancel,
		actions:     make(map[string]actionInfo),
		metrics:     make(map[string]int64),
		middlewares: nil,
	}

	defaults := []Option{
		Logger(logger.NewLogger(name, cfg.LogLevel)),
		Transport(transport.New(transport.Subject(subject))),
		Registry(registry.NewRegistry()),
		Selector(selector.NewSelector(registry.NewRegistry(), selector.SetStrategy(selector.RoundRobin))),
	}

	opts := append(defaults, extra...)
	s.opts = newOptions(opts...)

	if s.opts.Logger != nil {
		s.logger = s.opts.Logger
	}

	s.logger.Info("Name: %s | Version: %s | ID: %s", s.name, s.version, s.id)
	return s
}

// ID returns the service instance ID.
func (s *Service) ID() string               { return s.id }
func (s *Service) Name() string             { return s.name }
func (s *Service) Version() string          { return s.version }
func (s *Service) Options() Options         { return s.opts }
func (s *Service) Context() context.Context { return s.ctx }

// Config returns flattened config map.
func (s *Service) Config() map[string]string {
	return map[string]string{
		"service_name":       s.config.ServiceName,
		"bus_addr":           s.config.BusAddr,
		"log_level":          s.config.LogLevel,
		"port":               strconv.Itoa(s.config.Port),
		"dev_mode":           strconv.FormatBool(s.config.DevMode),
		"hc_memory_critical": strconv.FormatFloat(s.config.HCMemoryCriticalThreshold, 'f', -1, 64),
		"hc_memory_warning":  strconv.FormatFloat(s.config.HCMemoryWarningThreshold, 'f', -1, 64),
		"hc_load_critical":   strconv.FormatFloat(s.config.HCLoadCriticalThreshold, 'f', -1, 64),
		"hc_load_warning":    strconv.FormatFloat(s.config.HCLoadWarningThreshold, 'f', -1, 64),
	}
}

// Run starts the service and blocks until signal.
func (s *Service) Run() error {
	s.logger.Info("▶ starting %s %s", s.name, s.version)

	if err := s.start(); err != nil {
		return err
	}
	if s.opts.Hooks.OnStart != nil {
		s.opts.Hooks.OnStart()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	s.logger.Info("⏹ stopping %s %s", s.name, s.version)
	return s.Stop()
}

// Stop shuts down transport and deregisters service.
func (s *Service) Stop() error {
	s.cancel()

	if s.opts.Hooks.OnStop != nil {
		s.opts.Hooks.OnStop()
	}

	s.wg.Wait()
	if err := s.deregister(); err != nil {
		return err
	}
	return s.opts.Transport.Close()
}

// Init initializes and registers the service.
func (s *Service) Init(opts ...Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	if err := s.opts.Transport.Init(); err != nil {
		return err
	}

	s.opts.Transport.SetHandler(func(data []byte) error {
		msg := codec.NewMessage("")
		if err := codec.Unmarshal(data, msg); err != nil {
			return err
		}
		s.wg.Add(1)
		defer s.wg.Done()
		s.ServerHandler(msg)
		return nil
	})

	if err := s.opts.Registry.Init(); err != nil {
		return err
	}
	if err := s.opts.Selector.Init(); err != nil {
		return err
	}

	if s.opts.Router == nil {
		s.opts.Router = router.NewRouter(router.Name(s.name + "/" + s.version))
	}

	for name, info := range s.actions {
		wrapped := chainMiddlewares(info.handler, s.middlewares...)
		s.opts.Router.Add(&router.Node{
			ID:      name,
			Handler: s.prepareHandler(wrapped),
		})
	}

	s.announce()
	return nil
}

// start subscribes to transport and registers in registry.
func (s *Service) start() error {
	if err := s.register(); err != nil {
		return err
	}
	return s.opts.Transport.Subscribe()
}

// register adds the service to the registry.
func (s *Service) register() error {
	svc := &registry.Service{Name: s.name, Nodes: []*registry.Node{{ID: s.id}}}
	if s.opts.Router != nil {
		if err := s.opts.Router.Register(); err != nil {
			return err
		}
	}
	return s.opts.Registry.Register(svc)
}

// deregister removes the service from the registry.
func (s *Service) deregister() error {
	svc := &registry.Service{Name: s.name, Nodes: []*registry.Node{{ID: s.id}}}
	if err := s.opts.Registry.Deregister(svc); err != nil {
		return err
	}
	if s.opts.Router != nil {
		return s.opts.Router.Deregister()
	}
	return nil
}

// announce sends "file.register" with service schema info.
func (s *Service) announce() {
	payload := map[string]any{
		"service": s.name,
		"actions": s.ListActions(),
		"schemas": s.GetSchemas(),
	}
	data, err := codec.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal announce: %v", err)
		return
	}
	if err := s.opts.Transport.Publish("file.register", data); err != nil {
		s.logger.Error("failed to publish announce: %v", err)
		return
	}
	s.logger.Info("announced %d actions", len(s.actions))
}

// SubscribeTopic подписывается на указанный топик.
func (s *Service) SubscribeTopic(topic string, handler transport.MsgHandler) error {
	return s.opts.Transport.SubscribeTopic(topic, handler)
}

// SubscribePrefix подписывается на все топики с указанным префиксом.
func (s *Service) SubscribePrefix(prefix string, handler transport.MsgHandler) error {
	if s.opts.Transport == nil {
		return transport.ErrDisconnected
	}
	return s.opts.Transport.SubscribePrefix(prefix, handler)
}
