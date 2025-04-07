package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rskv-p/mini/pkg/x_log"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestServiceBasics(t *testing.T) {
	s := RunServerOnPort(-1)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Expected to connect to server, got %v", err)
	}
	defer nc.Close()

	log, _ := x_log.NewLogger()
	_ = log.Configure(x_log.OutputConsole, x_log.DebugLevel)

	// Handler stub
	doAdd := func(req Request) {
		type payload struct{ X, Y int }
		var p payload
		if err := json.Unmarshal(req.Data(), &p); err != nil {
			_ = req.Error("400", "invalid payload", nil)
			return
		}
		result := p.X + p.Y
		_ = req.RespondJSON(map[string]any{"sum": result})
		log.Debugw("handler finished", "subject", req.Subject())
	}

	cfg := Config{
		Name:        "math-service",
		Version:     "1.2.3",
		Description: "performs math operations",
		Logger:      log,
		Middleware: []Middleware{
			func(next Handler) Handler {
				return HandlerFunc(func(req Request) {
					log.Debugw("middleware hit", "subject", req.Subject())
					next.Handle(req)
				})
			},
		},
		Endpoint: &EndpointConfig{
			Subject: "math.add",
			Handler: HandlerFunc(doAdd),
		},
	}

	svc := AddService(nc, cfg)

	if err := svc.Start(); err != nil {
		t.Fatalf("service failed to start: %v", err)
	}

	resp, err := nc.Request("math.add", []byte(`{"x":2,"y":3}`), time.Second)
	if err != nil {
		t.Fatalf("Expected a response, got %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		t.Fatalf("Invalid JSON in response: %v", err)
	}
	if result["sum"] != float64(5) {
		t.Fatalf("Expected sum=5, got %v", result["sum"])
	}

	// Info check
	infoSubj, _ := ControlSubject(InfoVerb, "math-service", "")
	infoResp, err := nc.Request(infoSubj, nil, time.Second)
	if err != nil {
		t.Fatalf("Info request failed: %v", err)
	}
	var info Info
	_ = json.Unmarshal(infoResp.Data, &info)
	if info.Name != "math-service" {
		t.Fatalf("Unexpected info: %+v", info)
	}

	// Stats
	statsSubj, _ := ControlSubject(StatsVerb, "math-service", "")
	statsResp, err := nc.Request(statsSubj, nil, time.Second)
	if err != nil {
		t.Fatalf("Stats request failed: %v", err)
	}
	var stats Stats
	_ = json.Unmarshal(statsResp.Data, &stats)
	if stats.Endpoints[0].NumRequests != 1 {
		t.Fatalf("Expected 1 request, got %d", stats.Endpoints[0].NumRequests)
	}

	// Reset
	svc.Reset()
	if svc.Stats().Endpoints[0].NumRequests != 0 {
		t.Fatalf("Reset did not clear stats")
	}

	// Graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		_ = svc.Stop()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: svc.Stop() did not return")
	}
}

func TestAddService(t *testing.T) {
	log, _ := x_log.NewLogger()
	_ = log.Configure(x_log.OutputConsole, x_log.DebugLevel)

	handler := func(req Request) {
		_ = req.RespondJSON(map[string]string{"ok": "true"})
	}

	s := RunServerOnPort(-1)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	cfg := Config{
		Name:    "test_service",
		Version: "0.1.0",
		Logger:  log,
		Endpoint: &EndpointConfig{
			Subject:  "test.echo",
			Handler:  HandlerFunc(handler),
			Metadata: map[string]string{"x": "y"},
		},
		Metadata: map[string]string{"env": "test"},
	}

	svc := AddService(nc, cfg)
	if err := svc.Start(); err != nil {
		t.Fatalf("service start failed: %v", err)
	}
	defer svc.Stop()

	subj, _ := ControlSubject(PingVerb, "test_service", svc.Info().ID)
	msg, err := nc.Request(subj, nil, time.Second)
	if err != nil {
		t.Fatalf("ping request failed: %v", err)
	}

	var ping Ping
	_ = json.Unmarshal(msg.Data, &ping)

	if ping.Type != PingResponseType || ping.Name != "test_service" || ping.Version != "0.1.0" {
		t.Errorf("unexpected ping response: %+v", ping)
	}

	// Verify endpoint works
	resp, err := nc.Request("test.echo", []byte(`{}`), time.Second)
	if err != nil {
		t.Fatalf("handler request failed: %v", err)
	}

	var result map[string]string
	_ = json.Unmarshal(resp.Data, &result)
	if result["ok"] != "true" {
		t.Errorf("unexpected handler response: %+v", result)
	}
}

func RunServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}
