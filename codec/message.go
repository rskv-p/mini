package codec

import (
	"encoding/json"
	"errors"
)

// Ensure Message implements IMessage.
var _ IMessage = (*Message)(nil)

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
	Validate() error
}

// Message is the standard message format for transport.
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
// Constructors
// ----------------------------------------------------

// NewMessage creates a new empty message of a given type.
func NewMessage(t string) *Message {
	return &Message{
		Type:    t,
		Body:    make(map[string]any),
		Headers: make(map[string]string),
	}
}

// NewRequest creates a new request message for a given node and contextID.
func NewRequest(node, contextID string) *Message {
	return &Message{
		Type:      "request",
		Node:      node,
		ContextID: contextID,
		Body:      make(map[string]any),
		Headers:   make(map[string]string),
	}
}

// NewResponse creates a new response message with status code.
func NewResponse(contextID string, status int) *Message {
	return &Message{
		Type:       "response",
		ContextID:  contextID,
		StatusCode: status,
		Body:       make(map[string]any),
		Headers:    make(map[string]string),
	}
}

// NewJsonResponse creates a response message and precomputes RawBody.
func NewJsonResponse(contextID string, status int) *Message {
	m := NewResponse(contextID, status)
	_ = m.UpdateRawBody()
	return m
}

// Getters and setters

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

// Body access

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

// Raw body

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

// Result/Error

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

// Copy returns a deep copy of the message.
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

// Validate checks required fields.
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
