package core

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestRequestRespond(t *testing.T) {
	type x struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	tests := []struct {
		name             string
		respondData      any
		respondHeaders   Headers
		errDescription   string
		errCode          string
		errData          []byte
		expectedMessage  string
		expectedCode     string
		expectedResponse []byte
		withRespondError error
	}{
		{
			name:             "byte response",
			respondData:      []byte("OK"),
			expectedResponse: []byte("OK"),
		},
		{
			name:             "byte response, with headers",
			respondHeaders:   Headers{"key": []string{"value"}},
			respondData:      []byte("OK"),
			expectedResponse: []byte("OK"),
		},
		{
			name:             "byte response, connection closed",
			respondData:      []byte("OK"),
			withRespondError: ErrRespond,
		},
		{
			name:             "struct response",
			respondData:      x{"abc", 5},
			expectedResponse: []byte(`{"a":"abc","b":5}`),
		},
		{
			name:             "invalid response data",
			respondData:      func() {},
			withRespondError: ErrMarshalResponse,
		},
		{
			name:            "generic error",
			errDescription:  "oops",
			errCode:         "500",
			errData:         []byte("error!"),
			expectedMessage: "oops",
			expectedCode:    "500",
		},
		{
			name:            "generic error, with headers",
			respondHeaders:  Headers{"key": []string{"value"}},
			errDescription:  "oops",
			errCode:         "500",
			errData:         []byte("error!"),
			expectedMessage: "oops",
			expectedCode:    "500",
		},
		{
			name:            "error without response payload",
			errDescription:  "oops",
			errCode:         "500",
			expectedMessage: "oops",
			expectedCode:    "500",
		},
		{
			name:             "missing error code",
			errDescription:   "oops",
			withRespondError: ErrArgRequired,
		},
		{
			name:             "missing error description",
			errCode:          "500",
			withRespondError: ErrArgRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := RunServerOnPort(-1)
			defer s.Shutdown()

			nc, err := nats.Connect(s.ClientURL())
			if err != nil {
				t.Fatalf("Expected to connect to server, got %v", err)
			}
			defer nc.Close()

			handler := func(req Request) {
				if errors.Is(test.withRespondError, ErrRespond) {
					nc.Close()
					return
				}
				if val := req.Headers().Get("key"); val != "value" && len(test.respondHeaders) > 0 {
					t.Fatalf("Expected headers in request")
				}
				if !bytes.Equal(req.Data(), []byte("req")) {
					t.Fatalf("Invalid request data")
				}
				if test.errCode == "" && test.errDescription == "" {
					switch v := test.respondData.(type) {
					case []byte:
						err := req.Respond(v, WithHeaders(test.respondHeaders))
						if test.withRespondError != nil && !errors.Is(err, test.withRespondError) {
							t.Fatalf("Expected error: %v; got: %v", test.withRespondError, err)
						}
					default:
						err := req.RespondJSON(v, WithHeaders(test.respondHeaders))
						if test.withRespondError != nil && !errors.Is(err, test.withRespondError) {
							t.Fatalf("Expected error: %v; got: %v", test.withRespondError, err)
						}
					}
					return
				}
				err := req.Error(test.errCode, test.errDescription, test.errData, WithHeaders(test.respondHeaders))
				if test.withRespondError != nil && !errors.Is(err, test.withRespondError) {
					t.Fatalf("Expected error: %v; got: %v", test.withRespondError, err)
				}
			}

			svc := AddService(nc, Config{
				Name:    "CoolService",
				Version: "0.1.0",
				Endpoint: &EndpointConfig{
					Subject: "test.func",
					Handler: HandlerFunc(handler),
				},
			})
			if err := svc.Start(); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			defer svc.Stop()

			resp, err := nc.RequestMsg(&nats.Msg{
				Subject: "test.func",
				Data:    []byte("req"),
				Header:  nats.Header{"key": []string{"value"}},
			}, time.Second)

			if test.withRespondError != nil {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}

			if test.errCode != "" {
				if resp.Header.Get(ErrorCodeHeader) != test.expectedCode {
					t.Fatalf("Expected error code %q, got %q", test.expectedCode, resp.Header.Get(ErrorCodeHeader))
				}
				if resp.Header.Get(ErrorHeader) != test.expectedMessage {
					t.Fatalf("Expected error message %q, got %q", test.expectedMessage, resp.Header.Get(ErrorHeader))
				}
				return
			}

			if !bytes.Equal(resp.Data, test.expectedResponse) {
				t.Fatalf("Expected response: %q, got: %q", test.expectedResponse, resp.Data)
			}
			if !reflect.DeepEqual(Headers(resp.Header), test.respondHeaders) && len(test.respondHeaders) > 0 {
				t.Fatalf("Expected headers: %v, got: %v", test.respondHeaders, resp.Header)
			}
		})
	}
}
