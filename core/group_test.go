package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestGroups(t *testing.T) {
	tests := []struct {
		name             string
		endpointName     string
		groups           []string
		expectedEndpoint EndpointInfo
	}{
		{
			name:         "no groups",
			endpointName: "foo",
			expectedEndpoint: EndpointInfo{
				Name:       "foo",
				Subject:    "foo",
				QueueGroup: "q",
			},
		},
		{
			name:         "single group",
			endpointName: "foo",
			groups:       []string{"g1"},
			expectedEndpoint: EndpointInfo{
				Name:       "foo",
				Subject:    "g1.foo",
				QueueGroup: "q",
			},
		},
		{
			name:         "single empty group",
			endpointName: "foo",
			groups:       []string{""},
			expectedEndpoint: EndpointInfo{
				Name:       "foo",
				Subject:    "foo",
				QueueGroup: "q",
			},
		},
		{
			name:         "empty groups",
			endpointName: "foo",
			groups:       []string{"", "g1", ""},
			expectedEndpoint: EndpointInfo{
				Name:       "foo",
				Subject:    "g1.foo",
				QueueGroup: "q",
			},
		},
		{
			name:         "multiple groups",
			endpointName: "foo",
			groups:       []string{"g1", "g2", "g3"},
			expectedEndpoint: EndpointInfo{
				Name:       "foo",
				Subject:    "g1.g2.g3.foo",
				QueueGroup: "q",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := RunServerOnPort(-1)
			defer s.Shutdown()

			nc, err := nats.Connect(s.ClientURL())
			if err != nil {
				t.Fatalf("connect failed: %v", err)
			}
			defer nc.Close()

			svc := AddService(nc, Config{
				Name:        "test_service",
				Version:     "0.0.1",
				QueueGroup:  "q",
				Description: "test",
			})
			if err := svc.Start(); err != nil {
				t.Fatalf("start failed: %v", err)
			}
			defer svc.Stop()

			var group Group = svc
			for _, g := range test.groups {
				group = group.AddGroup(g)
			}

			err = group.AddEndpoint(test.endpointName, HandlerFunc(func(r Request) {}))
			if err != nil {
				t.Fatalf("AddEndpoint failed: %v", err)
			}

			time.Sleep(10 * time.Millisecond) // give time for endpoint to register

			info := svc.Info()
			if len(info.Endpoints) != 1 {
				t.Fatalf("expected 1 endpoint, got %d", len(info.Endpoints))
			}

			found := false
			for _, ep := range info.Endpoints {
				if ep.Name == test.expectedEndpoint.Name {
					if !reflect.DeepEqual(ep.Subject, test.expectedEndpoint.Subject) ||
						ep.QueueGroup != test.expectedEndpoint.QueueGroup {
						t.Fatalf("mismatch in endpoint: want %+v, got %+v", test.expectedEndpoint, ep)
					}
					found = true
				}
			}
			if !found {
				t.Fatalf("expected endpoint %q not found", test.expectedEndpoint.Name)
			}
		})
	}
}
