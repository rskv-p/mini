package runn_api

import "github.com/rskv-p/mini/servs/s_runn/runn_serv"

// OutgoingEvent represents an event sent to the client over WebSocket
type OutgoingEvent struct {
	Type    string             `json:"type"`    // Type of the event (e.g., "status_update")
	Process *runn_serv.Process `json:"process"` // The process associated with the event
}
