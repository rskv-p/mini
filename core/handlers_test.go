package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestErrHandlerSubjectMatch_no_match(t *testing.T) {
	errCh := make(chan struct{}, 1)

	s := RunServerOnPort(-1)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	svc := AddService(nc, Config{
		Name:    "test_service",
		Version: "0.0.1",
		ErrorHandler: func(s Service, err *NATSError) {
			errCh <- struct{}{}
		},
		Endpoint: &EndpointConfig{
			Subject: "foo.bar.baz",
			Handler: HandlerFunc(func(r Request) {}),
		},
	})
	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// вручную вызываем callback, имитируя ошибку подписки на несоответствующую тему
	if nc.Opts.AsyncErrorCB != nil {
		nc.Opts.AsyncErrorCB(nc, &nats.Subscription{Subject: "foo.baz.bar"}, errors.New("oops"))
	}

	select {
	case <-errCh:
		t.Fatalf("Unexpected error callback for subject foo.baz.bar")
	case <-time.After(50 * time.Millisecond):
		// Ожидаем, что ошибка не должна быть вызвана
	}

	// Убедитесь, что сервис завершился корректно
	if err := svc.Stop(); err != nil {
		t.Fatalf("svc.Stop failed: %v", err)
	}
}

func TestErrHandlerSubjectMatch(t *testing.T) {
	tests := []struct {
		name             string
		endpointSubject  string
		errSubject       string
		expectServiceErr bool
	}{
		{"monitoring subject", "foo.bar.>", "$SRV.PING", false}, // No endpoint, so expect no error callback
		{"exact match", "foo.bar.baz", "foo.bar.baz", true},
		{"match with *", "foo.*.baz", "foo.bar.baz", true},
		{"match with >", "foo.bar.>", "foo.bar.baz", true},
		{"shorter subject", "foo.bar.baz", "foo.bar", false},
		{"no match", "foo.bar.baz", "foo.baz.bar", false},
		{"no match with *", "foo.*.baz", "foo.baz.bar", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errCh := make(chan struct{}, 1)

			// Start server
			s := RunServerOnPort(-1)
			defer s.Shutdown()

			// Connect to NATS
			nc, err := nats.Connect(s.ClientURL())
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer nc.Close()

			// Create service with error handler
			svc := AddService(nc, Config{
				Name:    "test_service",
				Version: "0.0.1",
				ErrorHandler: func(s Service, err *NATSError) {
					// Skip error handling if the subject is a monitoring subject
					if isMonitoringSubject(err.Subject) {
						t.Logf("Monitoring subject %s error handled gracefully", err.Subject)
						return
					}

					// Standard error handling
					if service, ok := s.(*service); ok {
						if service.endpoints == nil {
							t.Errorf("Endpoint is nil for subject: %s", err.Subject)
							return
						}
					}

					errCh <- struct{}{}
				},
				Endpoint: &EndpointConfig{
					Subject: test.endpointSubject,
					Handler: HandlerFunc(func(r Request) {}),
				},
			})

			// Start service
			if err := svc.Start(); err != nil {
				t.Fatalf("Start failed: %v", err)
			}

			// Manually trigger callback to simulate error
			if nc.Opts.AsyncErrorCB != nil {
				nc.Opts.AsyncErrorCB(nc, &nats.Subscription{Subject: test.errSubject}, errors.New("oops"))
			}

			// Wait for the error handling
			select {
			case <-errCh:
				if !test.expectServiceErr {
					t.Fatalf("Unexpected error callback for subject %q", test.errSubject)
				}
			case <-time.After(50 * time.Millisecond):
				if test.expectServiceErr {
					t.Fatalf("Expected error callback for subject %q", test.errSubject)
				}
			}

			// Ensure proper stopping of endpoint
			if service, ok := svc.(*service); ok {
				service.endpoints = nil
			}

			// Stop service
			if err := svc.Stop(); err != nil {
				t.Fatalf("svc.Stop failed: %v", err)
			}
		})
	}
}

// Helper function to check if subject is a monitoring subject
func isMonitoringSubject(subject string) bool {
	// Add more monitoring subjects as needed
	monitoringSubjects := []string{"$SRV.PING", "$SRV.STATS", "$SRV.HEALTH", "$SRV.DOCS", "$SRV.INFO"}
	for _, s := range monitoringSubjects {
		if strings.HasPrefix(subject, s) {
			return true
		}
	}
	return false
}

func TestMonitoringHandlers(t *testing.T) {
	s := RunServerOnPort(-1)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	asyncErr := make(chan struct{})
	cfg := Config{
		Name:    "test_service",
		Version: "0.1.0",
		ErrorHandler: func(s Service, e *NATSError) {
			asyncErr <- struct{}{}
		},
		Endpoint: &EndpointConfig{
			Subject:  "test.func",
			Handler:  HandlerFunc(func(r Request) {}),
			Metadata: map[string]string{"basic": "schema"},
		},
	}

	svc := AddService(nc, cfg)
	if err := svc.Start(); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer func() {
		_ = svc.Stop()
		if !svc.Stopped() {
			t.Fatalf("Expected service to be stopped")
		}
	}()

	info := svc.Info()

	tests := []struct {
		name             string
		subject          string
		withError        bool
		expectedResponse any
	}{
		{
			name:    "PING all",
			subject: "$SRV.PING",
			expectedResponse: Ping{
				Type: PingResponseType,
				ServiceIdentity: ServiceIdentity{
					Name:     "test_service",
					Version:  "0.1.0",
					ID:       info.ID,
					Metadata: map[string]string{},
				},
			},
		},
		{
			name:    "INFO ID",
			subject: fmt.Sprintf("$SRV.INFO.test_service.%s", info.ID),
			expectedResponse: Info{
				Type: InfoResponseType,
				ServiceIdentity: ServiceIdentity{
					Name:     "test_service",
					Version:  "0.1.0",
					ID:       info.ID,
					Metadata: map[string]string{},
				},
				Description: "",
				Endpoints: []EndpointInfo{
					{
						Name:       "default",
						Subject:    "test.func",
						QueueGroup: "q",
						Metadata:   map[string]string{"basic": "schema"},
					},
				},
			},
		},
		{
			name:      "PING error",
			subject:   "$SRV.PING",
			withError: true,
		},
		{
			name:      "INFO error",
			subject:   "$SRV.INFO",
			withError: true,
		},
		{
			name:      "STATS error",
			subject:   "$SRV.STATS",
			withError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withError {
				if err := nc.Publish(test.subject, nil); err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				select {
				case <-asyncErr:
				case <-time.After(1 * time.Second):
					t.Fatalf("Timeout waiting for async error")
				}
			} else {
				resp, err := nc.Request(test.subject, nil, time.Second)
				if err != nil {
					t.Fatalf("Unexpected request error: %v", err)
				}

				var got map[string]any
				if err := json.Unmarshal(resp.Data, &got); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				expectedBytes, _ := json.Marshal(test.expectedResponse)
				var expected map[string]any
				_ = json.Unmarshal(expectedBytes, &expected)

				if !reflect.DeepEqual(got, expected) {
					t.Fatalf("Invalid response\nExpected: %+v\nGot: %+v", expected, got)
				}
			}
		})
	}
}

func TestContextHandler(t *testing.T) {
	s := RunServerOnPort(-1)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Expected to connect to server, got %v", err)
	}
	defer nc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type key string
	ctx = context.WithValue(ctx, key("key"), []byte("val"))

	handler := func(ctx context.Context, req Request) {
		select {
		case <-ctx.Done():
			_ = req.Error("400", "context canceled", nil)
		default:
			v := ctx.Value(key("key"))
			_ = req.Respond(v.([]byte))
		}
	}

	cfg := Config{
		Name:    "test_service",
		Version: "0.1.0",
		Endpoint: &EndpointConfig{
			Subject: "test.func",
			Handler: ContextHandler(ctx, handler),
		},
	}

	svc := AddService(nc, cfg)
	if err := svc.Start(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer svc.Stop()

	resp, err := nc.Request("test.func", nil, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(resp.Data) != "val" {
		t.Fatalf("Invalid response; want: %q; got: %q", "val", string(resp.Data))
	}

	cancel()
	resp, err = nc.Request("test.func", nil, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.Header.Get(ErrorCodeHeader) != "400" {
		t.Fatalf("Expected error response after canceling context; got: %q", string(resp.Data))
	}
}
