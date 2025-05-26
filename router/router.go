// file: mini/router/router.go
package router

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/constant"
)

var _ IRouter = (*Router)(nil)

// Handler processes an incoming IMessage with context and reply subject.
type Handler func(ctx context.Context, msg codec.IMessage, replyTo string) *Error

// HandlerWrapper wraps a Handler with middleware.
type HandlerWrapper func(Handler) Handler

// Error represents a handler error with status code.
type Error struct {
	StatusCode int
	Message    string
}

// Node represents a route registration.
type Node struct {
	ID                 string              `json:"id"`
	Handler            Handler             `json:"-"`
	ValidationRules    map[string][]string `json:"validation_rules,omitempty"`
	ValidationMessages map[string]string   `json:"validation_messages,omitempty"`
	Wrappers           []HandlerWrapper    `json:"-"`
}

// IAction defines a declarative router action.
type IAction interface {
	ID() string
	Handler() Handler
	Validation() (map[string][]string, map[string]string)
}

// IRouter defines routing methods.
type IRouter interface {
	Init(opts ...Option) error
	Routes() []*Node
	Add(*Node)
	AddMany([]*Node)
	RegisterActions([]IAction)
	RegisterActionsFromStructs(any)
	Dispatch(codec.IMessage) (Handler, error)
	Register() error
	Deregister() error
	GetOptions() Options
}

// Router implements IRouter.
type Router struct {
	routes map[string]*Node
	opts   Options
}

func NewRouter(opts ...Option) *Router {
	options := WithDefaults()
	for _, o := range opts {
		o(&options)
	}
	return &Router{
		routes: make(map[string]*Node),
		opts:   options,
	}
}

func (r *Router) Init(opts ...Option) error {
	for _, o := range opts {
		o(&r.opts)
	}
	return nil
}

func (r *Router) Routes() []*Node {
	list := make([]*Node, 0, len(r.routes))
	for _, n := range r.routes {
		list = append(list, n)
	}
	return list
}

func (r *Router) Add(n *Node) {
	if n == nil || n.ID == "" || n.Handler == nil {
		return
	}

	wrappers := append(r.opts.Wrappers, n.Wrappers...)
	n.Handler = wrapWithValidation(n, Wrap(n.Handler, wrappers))

	r.routes[n.ID] = n

	if r.opts.Logger != nil {
		r.opts.Logger.Debug("route registered: %s", n.ID)
	}
}

func (r *Router) AddMany(nodes []*Node) {
	for _, n := range nodes {
		r.Add(n)
	}
}

func (r *Router) RegisterActions(actions []IAction) {
	for _, act := range actions {
		if act == nil {
			continue
		}
		rules, msgs := act.Validation()
		r.Add(&Node{
			ID:                 act.ID(),
			Handler:            act.Handler(),
			ValidationRules:    rules,
			ValidationMessages: msgs,
		})
	}
}

func (r *Router) RegisterActionsFromStructs(list any) {
	val := reflect.ValueOf(list)
	if val.Kind() != reflect.Slice {
		log.Printf("[router] expected slice, got %T", list)
		return
	}
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		act, ok := item.(IAction)
		if !ok {
			log.Printf("[router] item at index %d does not implement IAction: %T", i, item)
			continue
		}
		rules, msgs := act.Validation()
		r.Add(&Node{
			ID:                 act.ID(),
			Handler:            act.Handler(),
			ValidationRules:    rules,
			ValidationMessages: msgs,
		})
	}
}

func (r *Router) Dispatch(msg codec.IMessage) (Handler, error) {
	if msg == nil {
		return nil, constant.ErrEmptyMessage
	}
	if msg.GetNode() == "" {
		return nil, constant.ErrInvalidPath
	}
	n, ok := r.routes[msg.GetNode()]
	if !ok {
		if r.opts.NotFound != nil {
			return func(ctx context.Context, msg codec.IMessage, replyTo string) *Error {
				return r.opts.NotFound(ctx, msg, replyTo)
			}, nil
		}
		return nil, constant.ErrNotFound
	}
	return r.wrapWithErrorHook(n.Handler), nil
}

func (r *Router) Register() error {
	return nil
}

func (r *Router) Deregister() error {
	r.routes = make(map[string]*Node)
	return nil
}

func (r *Router) GetOptions() Options {
	return r.opts
}

// ----------------------------------------------------
// Wrappers and validation
// ----------------------------------------------------

func Wrap(h Handler, wrappers []HandlerWrapper) Handler {
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}

func wrapWithValidation(n *Node, next Handler) Handler {
	if len(n.ValidationRules) == 0 {
		return next
	}
	return func(ctx context.Context, msg codec.IMessage, replyTo string) *Error {
		body := msg.GetBodyMap()
		for field, rules := range n.ValidationRules {
			val, exists := body[field]
			for _, rule := range rules {
				if err := validateRule(rule, val, exists, field); err != nil {
					return &Error{
						StatusCode: 400,
						Message:    validationMsg(n, field, rule, err.Error()),
					}
				}
			}
		}
		return next(ctx, msg, replyTo)
	}
}

func validateRule(rule string, val any, exists bool, field string) error {
	switch {
	case rule == "required":
		if !exists || val == nil || val == "" {
			return fmt.Errorf("field '%s' is required", field)
		}
	case strings.HasPrefix(rule, "min:"):
		return checkMin(val, strings.TrimPrefix(rule, "min:"))
	case strings.HasPrefix(rule, "max:"):
		return checkMax(val, strings.TrimPrefix(rule, "max:"))
	}
	return nil
}

func validationMsg(n *Node, field, rule, fallback string) string {
	if n == nil || n.ValidationMessages == nil {
		return fallback
	}
	// exact match: "name.min:3"
	if msg, ok := n.ValidationMessages[field+"."+rule]; ok {
		return msg
	}
	// rule prefix match: "name.min"
	if i := strings.Index(rule, ":"); i > 0 {
		if msg, ok := n.ValidationMessages[field+"."+rule[:i]]; ok {
			return msg
		}
	}
	if msg, ok := n.ValidationMessages[field]; ok {
		return msg
	}
	return fallback
}

func checkMin(val any, raw string) error {
	limit, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fmt.Errorf("invalid min value: %s", raw)
	}
	switch v := val.(type) {
	case float64:
		if v < limit {
			return fmt.Errorf("must be ≥ %v", limit)
		}
	case int:
		if float64(v) < limit {
			return fmt.Errorf("must be ≥ %v", limit)
		}
	case string:
		if float64(len(v)) < limit {
			return fmt.Errorf("length must be ≥ %v", limit)
		}
	}
	return nil
}

func checkMax(val any, raw string) error {
	limit, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fmt.Errorf("invalid max value: %s", raw)
	}
	switch v := val.(type) {
	case float64:
		if v > limit {
			return fmt.Errorf("must be ≤ %v", limit)
		}
	case int:
		if float64(v) > limit {
			return fmt.Errorf("must be ≤ %v", limit)
		}
	case string:
		if float64(len(v)) > limit {
			return fmt.Errorf("length must be ≤ %v", limit)
		}
	}
	return nil
}

func (r *Router) wrapWithErrorHook(h Handler) Handler {
	if r.opts.OnError == nil {
		return h
	}
	return func(ctx context.Context, msg codec.IMessage, replyTo string) *Error {
		err := h(ctx, msg, replyTo)
		if err != nil {
			r.opts.OnError(ctx, msg, err)
		}
		return err
	}
}
