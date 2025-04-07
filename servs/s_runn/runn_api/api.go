// servs/runn/runn_api/api.go
package runn_api

import "context"

type IRunn interface {
	List(ctx context.Context) ([]ServiceInfo, error)
	StartService(ctx context.Context, name string) error
	StopService(ctx context.Context, name string) error
	StopAllServices(ctx context.Context) error
}

type ServiceInfo struct {
	Name    string `json:"name"`    // Service name
	Running bool   `json:"running"` // Whether it's running
	Pid     int    `json:"pid"`     // OS process ID (if running)
	Uptime  int64  `json:"uptime"`  // Uptime in seconds
}

type ListRequest struct{}
type ListResponse struct {
	Services []ServiceInfo `json:"services"`
}

type StartRequest struct {
	Name string `json:"name"`
}
type StartResponse struct {
	Success bool `json:"success"`
}

type StopRequest struct {
	Name string `json:"name"`
}
type StopResponse struct {
	Success bool `json:"success"`
}

type StopAllRequest struct{}
type StopAllResponse struct {
	Stopped int `json:"stopped"`
}
