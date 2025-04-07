// servs/s_nats/nats_serv/service.go
package nats_serv

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rskv-p/mini/core"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/servs/s_nats/nats_api"
	"github.com/rskv-p/mini/servs/s_nats/nats_cfg"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type natsService struct {
	cfg  nats_cfg.NatsConfig
	log  x_log.Logger
	nc   *nats.Conn
	ns   *server.Server
	core core.Service
}

// New creates a new embedded NATS service.
func New(cfg nats_cfg.NatsConfig) core.Service {
	log, _ := x_log.NewLogger()

	return &natsService{
		cfg: cfg,
		log: log,
	}
}

// Init starts the embedded NATS server and core service.
func (s *natsService) Init() error {
	opts := &server.Options{
		Host:      s.cfg.Host,
		Port:      s.cfg.Port,
		JetStream: s.cfg.JetStream,
	}
	ns, err := server.NewServer(opts)
	if err != nil {
		return fmt.Errorf("nats-server init: %w", err)
	}

	s.ns = ns
	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		return fmt.Errorf("nats-server not ready")
	}

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return fmt.Errorf("nats client connect: %w", err)
	}
	s.nc = nc

	coreCfg := core.Config{
		Name:        s.cfg.Name,
		Version:     s.cfg.Version,
		Description: s.cfg.Description,
		QueueGroup:  s.cfg.QueueGroup,
		Logger:      s.log,
	}
	s.core = core.AddService(nc, coreCfg)
	if err := s.core.Init(); err != nil {
		return err
	}

	// Register echo endpoint
	return s.core.AddEndpoint(nats_api.SubjectEcho, core.HandlerFunc(func(req core.Request) {
		var in nats_api.EchoRequest
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			_ = req.Error("400", "invalid JSON", nil)
			return
		}
		_ = req.RespondJSON(nats_api.EchoResponse{Reply: in.Message})
	}), core.WithEndpointDoc(func(e *core.Endpoint) *core.EndpointDoc {
		return &core.EndpointDoc{
			Summary:     "Echo back the message",
			Request:     nats_api.EchoRequest{},
			Response:    nats_api.EchoResponse{},
			Description: "Simple echo endpoint for testing",
		}
	}))
}

// Start starts the core service.
func (s *natsService) Start() error {
	return s.core.Start()
}

// Stop stops the core service and embedded NATS.
func (s *natsService) Stop() error {
	_ = s.core.Stop()
	if s.nc != nil {
		s.nc.Close()
	}
	if s.ns != nil {
		s.ns.Shutdown()
	}
	return nil
}

func (s *natsService) AddEndpoint(name string, h core.Handler, opts ...core.EndpointOpt) error {
	return s.core.AddEndpoint(name, h, opts...)
}

func (s *natsService) AddGroup(name string, opts ...core.GroupOpt) core.Group {
	return s.core.AddGroup(name, opts...)
}

func (s *natsService) Info() core.Info {
	return s.core.Info()
}

func (s *natsService) Stats() core.Stats {
	return s.core.Stats()
}

func (s *natsService) Reset() {
	s.core.Reset()
}

func (s *natsService) Stopped() bool {
	return s.core.Stopped()
}
