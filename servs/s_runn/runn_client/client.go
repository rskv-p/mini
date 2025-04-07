// servs/runn/runn_client/client.go
package runn_client

import (
	"context"
	"time"

	"github.com/rskv-p/mini/servs/s_runn/runn_api"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
)

type LocalClient struct {
	svc *runn_serv.RunnService
}

// NewLocalClient returns a client for direct in-process control
func NewLocalClient(svc *runn_serv.RunnService) runn_api.IRunn {
	return &LocalClient{svc: svc}
}

func (c *LocalClient) List(ctx context.Context) ([]runn_api.ServiceInfo, error) {
	names := c.svc.ListLocal()
	result := make([]runn_api.ServiceInfo, 0, len(names))
	now := time.Now()

	for _, name := range names {
		info := runn_api.ServiceInfo{Name: name, Running: true}

		if proc := c.svc.GetProcess(name); proc != nil && proc.Cmd.Process != nil {
			info.Pid = proc.Pid
			info.Uptime = int64(now.Sub(proc.Started).Seconds())
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *LocalClient) StartService(ctx context.Context, name string) error {
	return c.svc.StartServiceLocal(name)
}

func (c *LocalClient) StopService(ctx context.Context, name string) error {
	return c.svc.StopServiceLocal(name)
}

func (c *LocalClient) StopAllServices(ctx context.Context) error {
	c.svc.StopAllServicesLocal()
	return nil
}
