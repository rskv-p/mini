package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestServiceStats(t *testing.T) {
	handler := func(r Request) {
		_ = r.Respond([]byte("ok")) // Basic handler response
	}

	tests := []struct {
		name          string
		config        Config
		expectedStats map[string]any
	}{
		{
			name: "stats handler", // Test without custom stats handler
			config: Config{
				Name:    "test_service",
				Version: "0.1.0",
			},
		},
		{
			name: "with stats handler", // Test with custom stats handler
			config: Config{
				Name:    "test_service",
				Version: "0.1.0",
				StatsHandler: func(e *Endpoint) any {
					return map[string]any{
						"key": "val", // Return custom stats
					}
				},
			},
			expectedStats: map[string]any{
				"key": "val", // Expected custom stats
			},
		},
		{
			name: "with default endpoint", // Test with default endpoint
			config: Config{
				Name:    "test_service",
				Version: "0.1.0",
				Endpoint: &EndpointConfig{
					Subject:  "test.func",
					Handler:  HandlerFunc(handler),
					Metadata: map[string]string{"test": "value"},
				},
			},
		},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := RunServerOnPort(-1) // Start server
			defer s.Shutdown()

			nc, err := nats.Connect(s.ClientURL()) // Connect to NATS server
			if err != nil {
				t.Fatalf("Expected to connect to server, got %v", err)
			}
			defer nc.Close()

			svc := AddService(nc, test.config) // Add service
			if err := svc.Start(); err != nil {
				t.Fatalf("Failed to start service: %v", err)
			}
			defer svc.Stop()

			if test.config.Endpoint == nil { // Add endpoint if not provided
				opts := []EndpointOpt{WithEndpointSubject("test.func")}
				if err := svc.AddEndpoint("func", HandlerFunc(handler), opts...); err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			// Send valid requests
			for i := 0; i < 10; i++ {
				if _, err := nc.Request("test.func", []byte("msg"), time.Second); err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			// Send invalid request (no reply subject)
			if err := nc.Publish("test.func", []byte("err")); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			time.Sleep(20 * time.Millisecond)

			// Request stats
			info := svc.Info()
			statsSubj := fmt.Sprintf("$SRV.STATS.test_service.%s", info.ID)
			resp, err := nc.Request(statsSubj, nil, time.Second)
			if err != nil {
				t.Fatalf("Failed to request stats: %v", err)
			}

			// Unmarshal stats response
			var stats Stats
			if err := json.Unmarshal(resp.Data, &stats); err != nil {
				t.Fatalf("Invalid JSON stats: %v", err)
			}

			// Validate stats
			if len(stats.Endpoints) != 1 {
				t.Fatalf("Expected 1 endpoint; got %d", len(stats.Endpoints))
			}

			ep := stats.Endpoints[0]
			expectedName := "default"
			if test.config.Endpoint == nil {
				expectedName = "func" // Set expected name if no default endpoint
			}
			if ep.Name != expectedName {
				t.Errorf("Invalid endpoint name; want: %q; got: %q", expectedName, ep.Name)
			}
			if ep.Subject != "test.func" {
				t.Errorf("Invalid subject; want: test.func; got: %s", ep.Subject)
			}
			if ep.NumRequests != 11 {
				t.Errorf("Unexpected NumRequests; want: 11; got: %d", ep.NumRequests)
			}
			if ep.NumErrors != 1 {
				t.Errorf("Unexpected NumErrors; want: 1; got: %d", ep.NumErrors)
			}
			if ep.AverageProcessingTime == 0 {
				t.Errorf("Expected non-zero AverageProcessingTime")
			}
			if ep.ProcessingTime == 0 {
				t.Errorf("Expected non-zero ProcessingTime")
			}
			if stats.Started.IsZero() {
				t.Errorf("Expected non-zero service start time")
			}
			if stats.Type != StatsResponseType {
				t.Errorf("Invalid stats type; want: %s; got: %s", StatsResponseType, stats.Type)
			}
			if stats.Name != info.Name {
				t.Errorf("Unexpected Name; want: %s; got: %s", info.Name, stats.Name)
			}
			if stats.ID != info.ID {
				t.Errorf("Unexpected ID; want: %s; got: %s", info.ID, stats.ID)
			}

			// Validate custom stats data
			if test.expectedStats != nil {
				var data map[string]any
				if err := json.Unmarshal(ep.Data, &data); err != nil {
					t.Fatalf("Invalid stats data: %v", err)
				}
				if !reflect.DeepEqual(data, test.expectedStats) {
					t.Fatalf("Stats handler mismatch; want: %+v; got: %+v", test.expectedStats, data)
				}
			}
		})
	}
}
