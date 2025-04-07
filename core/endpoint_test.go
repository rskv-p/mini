package core

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestAddEndpoint_Concurrency(t *testing.T) {
	// Run server on a random port and defer shutdown
	s := RunServerOnPort(-1)
	defer s.Shutdown()

	// Connect to NATS server
	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Expected to connect to server, got %v", err)
	}
	defer nc.Close()

	// Create context for handler
	ctx := context.Background()

	// Define handler function
	handler := func(ctx context.Context, req Request) {
		_ = req.RespondJSON(map[string]any{"hello": "world"})
	}

	// Add service and start it
	svc := AddService(nc, Config{
		Name:    "test_service",
		Version: "0.1.0",
	})
	if err := svc.Start(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer svc.Stop()

	// Channel and wait group for concurrency tests
	res := make(chan error, 10)
	wg := sync.WaitGroup{}
	wg.Add(10)

	// Add endpoints concurrently
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			// Use a unique endpoint name for each iteration
			endpointName := fmt.Sprintf("test%d", i)
			err := svc.AddEndpoint(endpointName, ContextHandler(ctx, handler))
			res <- err
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check for errors in adding endpoints
	for range make([]int, 10) {
		select {
		case err := <-res:
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for endpoint to be added")
		}
	}

	// Check endpoint count (including default endpoint)
	expectedCount := 10
	if cfg := svc.(*service).Config.Endpoint; cfg != nil {
		expectedCount++
	}
	if len(svc.Info().Endpoints) != expectedCount {
		t.Fatalf("Expected %d endpoints, got: %d", expectedCount, len(svc.Info().Endpoints))
	}
}
