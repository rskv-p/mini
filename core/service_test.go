package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/rskv-p/mini/pkg/x_log"
)

func TestServiceBasics(t *testing.T) {
	// Start the NATS server
	s := RunServerOnPort(-1)
	defer s.Shutdown()

	// Connect to the NATS server
	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Expected to connect to server, got %v", err)
	}
	defer nc.Close()

	// Handler stub for adding numbers
	doAdd := func(req Request) {
		type payload struct{ X, Y int }
		var p payload
		if err := json.Unmarshal(req.Data(), &p); err != nil {
			_ = req.Error("400", "invalid payload", nil)
			return
		}
		result := p.X + p.Y
		_ = req.RespondJSON(map[string]any{"sum": result})
		x_log.Debug().Str("subject", req.Subject()).Msg("handler finished")
	}

	// Configuring the service
	cfg := Config{
		Name:        "math-service",
		Version:     "1.2.3",
		Description: "performs math operations",
		Middleware: []Middleware{
			func(next Handler) Handler {
				return HandlerFunc(func(req Request) {
					x_log.Debug().Str("subject", req.Subject()).Msg("middleware hit")
					next.Handle(req)
				})
			},
		},
		Endpoint: &EndpointConfig{
			Subject: "math.add",
			Handler: HandlerFunc(doAdd),
		},
	}

	// Add the service to NATS
	svc := AddService(nc, cfg)

	// Start the service
	if err := svc.Start(); err != nil {
		t.Fatalf("service failed to start: %v", err)
	}

	// Send a request to the service
	resp, err := nc.Request("math.add", []byte(`{"x":2,"y":3}`), time.Second)
	if err != nil {
		t.Fatalf("Expected a response, got %v", err)
	}

	// Parse the response
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

	// Stats check
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

	// Reset service stats
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
		// Service stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: svc.Stop() did not return")
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
