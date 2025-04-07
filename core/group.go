package core

import (
	"fmt"
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

	if g.service.Logger != nil {
		g.service.Logger.Infow("created group",
			"parent", g.prefix,
			"name", name,
			"prefix", prefix,
			"queue_group", qg,
			"queue_disabled", noQ,
		)
	}

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
	options.logger = g.service.Logger // проброс логгера

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			if g.service.Logger != nil {
				g.service.Logger.Errorw("invalid endpoint option",
					"name", name,
					"err", err,
				)
			}
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

	if g.service.Logger != nil {
		g.service.Logger.Infow("adding endpoint to group",
			"name", name,
			"subject", endpointSubject,
			"queue_group", qg,
			"queue_disabled", noQ,
			"group_prefix", g.prefix,
		)
	}

	return addEndpoint(g.service, name, endpointSubject, handler, options.metadata, qg, noQ, options.logger)
}
