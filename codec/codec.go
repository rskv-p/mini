// file: arc/service/codec/codec.go
package codec

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// IMessage defines the contract for transport messages.
type IMessage interface {
	GetType() string
	SetType(string)
	GetNode() string
	SetNode(string)

	GetContextID() string
	SetContextID(string)

	GetReplyTo() string
	SetReplyTo(string)

	GetHeaders() map[string]string
	SetHeader(key, value string)
	GetHeader(key string) string

	GetBodyMap() map[string]any
	SetBody(obj any)
	Get(key string) (any, bool)
	Set(key string, val any)
	GetString(key string) string
	GetInt(key string) int64
	GetFloat(key string) float64
	GetBool(key string) bool

	GetRawBody() []byte
	UpdateRawBody() error

	SetResult(value any)
	GetResult(target any) error
	SetError(err error)
	GetError() string
	HasError() bool

	Copy() IMessage
}

var _ IMessage = (*Message)(nil)

type Message struct {
	Type       string            `json:"type,omitempty"`
	Node       string            `json:"node,omitempty"`
	ContextID  string            `json:"contextID,omitempty"`
	ReplyTo    string            `json:"replyTo,omitempty"`
	Headers    map[string]string `json:"header,omitempty"`
	Body       map[string]any    `json:"body,omitempty"`
	RawBody    []byte            `json:"rawBody,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
}

// ----------------------------------------------------
// IMessage Implementation
// ----------------------------------------------------

func (m *Message) GetType() string        { return m.Type }
func (m *Message) SetType(t string)       { m.Type = t }
func (m *Message) GetNode() string        { return m.Node }
func (m *Message) SetNode(n string)       { m.Node = n }
func (m *Message) GetContextID() string   { return m.ContextID }
func (m *Message) SetContextID(id string) { m.ContextID = id }
func (m *Message) GetReplyTo() string     { return m.ReplyTo }
func (m *Message) SetReplyTo(r string)    { m.ReplyTo = r }
func (m *Message) GetHeaders() map[string]string {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	return m.Headers
}
func (m *Message) SetHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}
func (m *Message) GetHeader(key string) string {
	if m.Headers == nil {
		return ""
	}
	return m.Headers[key]
}
func (m *Message) GetBodyMap() map[string]any {
	if m.Body == nil {
		m.Body = make(map[string]any)
	}
	return m.Body
}

func (m *Message) Set(key string, value any) {
	m.GetBodyMap()[key] = value
	m.RawBody = nil
}

func (m *Message) Get(key string) (any, bool) {
	v, ok := m.GetBodyMap()[key]
	return v, ok
}

func (m *Message) GetString(key string) string {
	v, _ := m.Get(key)
	s, _ := toString(v)
	return s
}

func (m *Message) GetInt(key string) int64 {
	v, _ := m.Get(key)
	i, _ := toInt64(v)
	return i
}

func (m *Message) GetFloat(key string) float64 {
	v, _ := m.Get(key)
	f, _ := toFloat64(v)
	return f
}

func (m *Message) GetBool(key string) bool {
	v, _ := m.Get(key)
	b, _ := toBool(v)
	return b
}

func (m *Message) SetBody(obj any) {
	if obj == nil {
		m.Body, m.RawBody = nil, nil
		return
	}
	b, err := json.Marshal(obj)
	if err != nil {
		m.Body, m.RawBody = nil, nil
		return
	}
	var body map[string]any
	if err := json.Unmarshal(b, &body); err != nil {
		m.Body, m.RawBody = nil, nil
		return
	}
	m.Body = body
	m.RawBody = nil
}

func (m *Message) GetRawBody() []byte {
	if m.RawBody == nil && len(m.Body) > 0 {
		_ = m.UpdateRawBody()
	}
	return m.RawBody
}

func (m *Message) UpdateRawBody() error {
	b, err := json.Marshal(m.GetBodyMap())
	if err != nil {
		return err
	}
	m.RawBody = b
	return nil
}

func (m *Message) SetResult(value any) {
	m.Set("result", value)
}

func (m *Message) GetResult(target any) error {
	raw, ok := m.GetBodyMap()["result"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func (m *Message) SetError(err error) {
	if err != nil {
		m.Set("error", err.Error())
	} else {
		delete(m.GetBodyMap(), "error")
	}
}

func (m *Message) GetError() string {
	return m.GetString("error")
}

func (m *Message) HasError() bool {
	err, ok := m.Body["error"]
	s, _ := toString(err)
	return ok && s != ""
}

func (m *Message) Copy() IMessage {
	if m == nil {
		return nil
	}
	clone := *m
	clone.Body = make(map[string]any, len(m.Body))
	for k, v := range m.Body {
		clone.Body[k] = v
	}
	clone.Headers = make(map[string]string, len(m.Headers))
	for k, v := range m.Headers {
		clone.Headers[k] = v
	}
	if m.RawBody != nil {
		clone.RawBody = make([]byte, len(m.RawBody))
		copy(clone.RawBody, m.RawBody)
	}
	return &clone
}

// ----------------------------------------------------
// Constructors
// ----------------------------------------------------

func NewMessage(t string) *Message {
	return &Message{Type: t, Body: map[string]any{}, Headers: map[string]string{}}
}

func NewRequest(node, contextID string) *Message {
	return &Message{
		Type:      "request",
		Node:      node,
		ContextID: contextID,
		Body:      map[string]any{},
		Headers:   map[string]string{},
	}
}

func NewResponse(contextID string, status int) *Message {
	return &Message{
		Type:       "response",
		ContextID:  contextID,
		StatusCode: status,
		Body:       map[string]any{},
		Headers:    map[string]string{},
	}
}

func NewJsonResponse(contextID string, status int) *Message {
	m := NewResponse(contextID, status)
	_ = m.UpdateRawBody()
	return m
}

func (m *Message) Validate() error {
	if m.Type == "" {
		return errors.New("missing Type")
	}
	if m.Type == "request" && m.Node == "" {
		return errors.New("missing Node for request")
	}
	if m.ContextID == "" {
		return errors.New("missing ContextID")
	}
	return nil
}

// ----------------------------------------------------
// Type conversions
// ----------------------------------------------------

func toString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case []byte:
		return string(x), true
	case int:
		return strconv.Itoa(x), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	default:
		b, err := json.Marshal(x)
		return string(b), err == nil
	}
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		return i, err == nil
	}
	return 0, false
}

func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

func toBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		b, err := strconv.ParseBool(x)
		return b, err == nil
	}
	return false, false
}

// ----------------------------------------------------
// JSON helpers
// ----------------------------------------------------

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal error: %v", err))
	}
	return b
}
