// file: arc/service/context/context.go
package context

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

var _ IContext = (*Context)(nil)

// ----------------------------------------------------
// Conversation and context manager
// ----------------------------------------------------

// Conversation represents a request/response session.
type Conversation struct {
	ID        string        // unique session ID
	Request   string        // reply-to topic
	done      chan struct{} // completion signal
	CreatedAt time.Time     // when it was added
	ClosedAt  time.Time     // when it was closed (Done)
}

// IContext manages Conversation lifecycles.
type IContext interface {
	Add(*Conversation) string
	With(id string, conv *Conversation)
	Get(id string) *Conversation
	Delete(id string)
	Done(id string)
	Wait(id string)
	WaitTimeout(id string, timeout time.Duration) bool
	WaitContext(id string, ctx context.Context) bool
	Has(id string) bool
	Count() int
	All() []*Conversation
	Range(func(id string, conv *Conversation) bool)
	Reset()
	SetAutoDelete(bool)
	SetHooks(onAdd func(*Conversation), onDelete func(*Conversation))
}

// Context implements IContext using sync.Map.
type Context struct {
	pool       sync.Map
	autoDelete bool
	onAdd      func(*Conversation)
	onDelete   func(*Conversation)
}

// NewContext returns a new Context manager.
func NewContext() IContext {
	return &Context{}
}

// SetHooks sets optional hooks on add/delete.
func (m *Context) SetHooks(onAdd func(*Conversation), onDelete func(*Conversation)) {
	m.onAdd = onAdd
	m.onDelete = onDelete
}

// Add stores conv and generates ID if missing.
func (m *Context) Add(conv *Conversation) string {
	if conv == nil {
		return ""
	}
	if conv.ID == "" || !isUUID(conv.ID) {
		conv.ID = uuid.NewString()
	}
	conv.done = make(chan struct{})
	conv.CreatedAt = time.Now()
	m.pool.Store(conv.ID, conv)
	if m.onAdd != nil {
		m.onAdd(conv)
	}
	return conv.ID
}

// With stores conv under explicit ID.
func (m *Context) With(id string, conv *Conversation) {
	if conv == nil || id == "" {
		return
	}
	conv.ID = id
	if conv.done == nil {
		conv.done = make(chan struct{})
	}
	conv.CreatedAt = time.Now()
	m.pool.Store(id, conv)
	if m.onAdd != nil {
		m.onAdd(conv)
	}
}

// Get retrieves a Conversation by ID.
func (m *Context) Get(id string) *Conversation {
	if val, ok := m.pool.Load(id); ok {
		return val.(*Conversation)
	}
	return nil
}

// Has returns true if ID exists.
func (m *Context) Has(id string) bool {
	_, ok := m.pool.Load(id)
	return ok
}

// Delete removes a Conversation.
func (m *Context) Delete(id string) {
	if conv := m.Get(id); conv != nil && m.onDelete != nil {
		m.onDelete(conv)
	}
	m.pool.Delete(id)
}

// Done signals completion for ID.
func (m *Context) Done(id string) {
	if conv := m.Get(id); conv != nil {
		select {
		case <-conv.done:
			// already closed
		default:
			close(conv.done)
			conv.ClosedAt = time.Now()
			if m.autoDelete {
				m.Delete(id)
			}
		}
	}
}

// Wait blocks until Done is called.
func (m *Context) Wait(id string) {
	if conv := m.Get(id); conv != nil && conv.done != nil {
		<-conv.done
	}
}

// WaitTimeout blocks until Done or timeout.
func (m *Context) WaitTimeout(id string, timeout time.Duration) bool {
	if conv := m.Get(id); conv != nil && conv.done != nil {
		select {
		case <-conv.done:
			return true
		case <-time.After(timeout):
			return false
		}
	}
	return false
}

// WaitContext blocks until Done or ctx is canceled.
func (m *Context) WaitContext(id string, ctx context.Context) bool {
	if conv := m.Get(id); conv != nil && conv.done != nil {
		select {
		case <-conv.done:
			return true
		case <-ctx.Done():
			return false
		}
	}
	return false
}

// Count returns number of sessions.
func (m *Context) Count() int {
	count := 0
	m.pool.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// All returns all conversations as a slice.
func (m *Context) All() []*Conversation {
	var out []*Conversation
	m.Range(func(_ string, conv *Conversation) bool {
		out = append(out, conv)
		return true
	})
	return out
}

// Range calls f for each session.
func (m *Context) Range(f func(id string, conv *Conversation) bool) {
	m.pool.Range(func(k, v any) bool {
		return f(k.(string), v.(*Conversation))
	})
}

// Reset clears all state (for tests).
func (m *Context) Reset() {
	m.pool = sync.Map{}
	m.onAdd = nil
	m.onDelete = nil
}

// SetAutoDelete enables removal after Done().
func (m *Context) SetAutoDelete(enable bool) {
	m.autoDelete = enable
}

// isUUID checks if str is a valid UUID.
func isUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}
