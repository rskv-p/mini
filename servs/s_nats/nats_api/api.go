// servs/s_nats/nats_api/api.go
package nats_api

const (
	SubjectEcho = "echo" // Example subject
)

// EchoRequest is a placeholder for request payload
type EchoRequest struct {
	Message string `json:"message"`
}

// EchoResponse is a placeholder for response payload
type EchoResponse struct {
	Reply string `json:"reply"`
}
