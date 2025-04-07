// servs/s_nats/s_nats_conf/config.go
package nats_cfg

type NatsConfig struct {
	Name        string `json:"name"`        // Service name
	Version     string `json:"version"`     // Service version
	Description string `json:"description"` // Description
	QueueGroup  string `json:"queue_group"` // Queue group for endpoints

	Host string `json:"host"` // NATS bind host
	Port int    `json:"port"` // NATS bind port

	JetStream bool `json:"jetstream"` // Enable JetStream
}
