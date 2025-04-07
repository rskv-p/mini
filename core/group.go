package core

import (
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"
)

// Group defines a namespace for endpoints.
type Group interface {
	AddGroup(string, ...GroupOpt) Group
	AddEndpoint(string, Handler, ...EndpointOpt) error
}

type group struct {
	service            *service
	prefix             string
	queueGroup         string
	queueGroupDisabled bool
}

// AddGroup creates a nested group with prefixed subject.
func (g *group) AddGroup(name string, opts ...GroupOpt) Group {
	var o groupOpts
	for _, opt := range opts {
		opt(&o)
	}

	qg, noQ := resolveQueueGroup(
		o.queueGroup, g.queueGroup,
		o.qgDisabled, g.queueGroupDisabled,
	)

	var parts []string
	if g.prefix != "" {
		parts = append(parts, g.prefix)
	}
	if name != "" {
		parts = append(parts, name)
	}

	prefix := joinParts(parts)

	// Логирование через глобальный логгер x_log
	x_log.Info().Str("parent", g.prefix).
		Str("name", name).
		Str("prefix", prefix).
		Str("queue_group", qg).
		Bool("queue_disabled", noQ).
		Msg("created group")

	return &group{
		service:            g.service,
		prefix:             prefix,
		queueGroup:         qg,
		queueGroupDisabled: noQ,
	}
}

// AddEndpoint registers a new endpoint with group prefix.
func (g *group) AddEndpoint(name string, handler Handler, opts ...EndpointOpt) error {
	var options endpointOpts

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			x_log.Error().Str("name", name).Err(err).
				Msg("invalid endpoint option")
			return err
		}
	}

	subject := options.subject
	if subject == "" {
		subject = name
	}

	endpointSubject := subject
	if g.prefix != "" {
		endpointSubject = fmt.Sprintf("%s.%s", g.prefix, subject)
	}

	qg, noQ := resolveQueueGroup(
		options.queueGroup, g.queueGroup,
		options.qgDisabled, g.queueGroupDisabled,
	)

	// Логирование через глобальный логгер x_log
	x_log.Info().Str("name", name).
		Str("subject", endpointSubject).
		Str("queue_group", qg).
		Bool("queue_disabled", noQ).
		Str("group_prefix", g.prefix).
		Msg("adding endpoint to group")

	return addEndpoint(g.service, name, endpointSubject, handler, options.metadata, qg, noQ)
}
