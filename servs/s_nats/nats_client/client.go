// servs/s_nats/nats_client/client.go
package nats_client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rskv-p/mini/core"
	"github.com/rskv-p/mini/servs/s_nats/nats_api"

	"github.com/nats-io/nats.go"
)

type Client interface {
	Echo(ctx context.Context, msg string) (string, error)
	Ping(ctx context.Context) error
	Stats(ctx context.Context) (core.Stats, error)
	Info(ctx context.Context) (core.Info, error)
}

type client struct {
	nc      *nats.Conn
	target  string // имя сервиса
	timeout time.Duration
}

// New returns a new Client for a target service
func New(nc *nats.Conn, target string) Client {
	return &client{
		nc:      nc,
		target:  target,
		timeout: 2 * time.Second,
	}
}

func (c *client) Echo(ctx context.Context, msg string) (string, error) {
	req := nats_api.EchoRequest{Message: msg}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := c.request(ctx, nats_api.SubjectEcho, data)
	if err != nil {
		return "", err
	}

	var out nats_api.EchoResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return "", err
	}

	return out.Reply, nil
}

func (c *client) Ping(ctx context.Context) error {
	subj, err := core.ControlSubject(core.PingVerb, c.target, "")
	if err != nil {
		return err
	}
	_, err = c.request(ctx, subj, nil)
	return err
}

func (c *client) Info(ctx context.Context) (core.Info, error) {
	var info core.Info
	subj, err := core.ControlSubject(core.InfoVerb, c.target, "")
	if err != nil {
		return info, err
	}

	msg, err := c.request(ctx, subj, nil)
	if err != nil {
		return info, err
	}

	if err := json.Unmarshal(msg.Data, &info); err != nil {
		return info, err
	}

	return info, nil
}

func (c *client) Stats(ctx context.Context) (core.Stats, error) {
	var stats core.Stats
	subj, err := core.ControlSubject(core.StatsVerb, c.target, "")
	if err != nil {
		return stats, err
	}

	msg, err := c.request(ctx, subj, nil)
	if err != nil {
		return stats, err
	}

	if err := json.Unmarshal(msg.Data, &stats); err != nil {
		return stats, err
	}

	return stats, nil
}

// request wraps nats request with context
func (c *client) request(ctx context.Context, subject string, data []byte) (*nats.Msg, error) {
	return c.nc.RequestWithContext(ctx, subject, data)
}
