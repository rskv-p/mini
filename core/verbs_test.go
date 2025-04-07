package core

import (
	"errors"
	"testing"
)

func TestControlSubject(t *testing.T) {
	// Test cases for ControlSubject function
	tests := []struct {
		name            string
		verb            Verb
		srvName         string
		id              string
		expectedSubject string
		withError       error
	}{
		{
			name:            "PING ALL", // Test for PING verb with no service name
			verb:            PingVerb,
			expectedSubject: "$SRV.PING", // Expected subject
		},
		{
			name:            "PING name", // Test for PING verb with service name
			verb:            PingVerb,
			srvName:         "test",
			expectedSubject: "$SRV.PING.test", // Expected subject with service name
		},
		{
			name:            "PING id", // Test for PING verb with service name and ID
			verb:            PingVerb,
			srvName:         "test",
			id:              "123",
			expectedSubject: "$SRV.PING.test.123", // Expected subject with service name and ID
		},
		{
			name:      "invalid verb", // Test for an unsupported verb
			verb:      Verb(100),
			withError: ErrVerbNotSupported, // Expected error for invalid verb
		},
		{
			name:      "name not provided", // Test for missing service name
			verb:      PingVerb,
			srvName:   "",
			id:        "123",
			withError: ErrServiceNameRequired, // Expected error for missing service name
		},
	}

	// Run each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ControlSubject(test.verb, test.srvName, test.id)
			if test.withError != nil {
				// Check if the expected error is returned
				if !errors.Is(err, test.withError) {
					t.Fatalf("Expected error: %v; got: %v", test.withError, err)
				}
				return
			}
			// Check if unexpected error occurred
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Check if the subject matches the expected value
			if res != test.expectedSubject {
				t.Errorf("Invalid subject; want: %q; got: %q", test.expectedSubject, res)
			}
		})
	}
}
