package context_test

import (
	dcon "context"
	"testing"
	"time"

	"github.com/rskv-p/mini/context"
	"github.com/stretchr/testify/assert"
)

func TestContext_AddGetDone(t *testing.T) {
	m := context.NewContext()

	conv := &context.Conversation{Request: "reply.test"}
	id := m.Add(conv)
	assert.NotEmpty(t, id)

	got := m.Get(id)
	assert.Equal(t, conv, got)
	assert.True(t, m.Has(id))

	done := make(chan bool)
	go func() {
		m.Wait(id)
		done <- true
	}()

	time.Sleep(10 * time.Millisecond)
	m.Done(id)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("wait did not unblock")
	}
}

func TestContext_AutoDelete(t *testing.T) {
	m := context.NewContext()
	m.SetAutoDelete(true)

	id := m.Add(&context.Conversation{})
	m.Done(id)

	time.Sleep(10 * time.Millisecond)
	assert.False(t, m.Has(id))
}

func TestContext_Delete(t *testing.T) {
	m := context.NewContext()

	var deleted bool
	m.SetHooks(nil, func(c *context.Conversation) {
		deleted = true
	})

	id := m.Add(&context.Conversation{})
	m.Delete(id)

	assert.False(t, m.Has(id))
	assert.True(t, deleted)
}

func TestContext_WaitTimeout(t *testing.T) {
	m := context.NewContext()
	id := m.Add(&context.Conversation{})

	ok := m.WaitTimeout(id, 10*time.Millisecond)
	assert.False(t, ok)

	m.Done(id)
	ok = m.WaitTimeout(id, 10*time.Millisecond)
	assert.True(t, ok)
}

func TestContext_WaitContext(t *testing.T) {
	m := context.NewContext()
	id := m.Add(&context.Conversation{})

	ctx, cancel := dcon.WithTimeout(dcon.Background(), 10*time.Millisecond)
	defer cancel()

	ok := m.WaitContext(id, ctx)
	assert.False(t, ok)

	m.Done(id)
	ok = m.WaitContext(id, dcon.Background())
	assert.True(t, ok)
}

func TestContext_WithExplicitID(t *testing.T) {
	m := context.NewContext()
	conv := &context.Conversation{}
	m.With("my-id", conv)

	assert.Equal(t, "my-id", conv.ID)
	assert.True(t, m.Has("my-id"))
}

func TestContext_Range_All_Count_Reset(t *testing.T) {
	m := context.NewContext()
	m.Add(&context.Conversation{})
	m.Add(&context.Conversation{})

	assert.Equal(t, 2, m.Count())
	assert.Len(t, m.All(), 2)

	calls := 0
	m.Range(func(id string, conv *context.Conversation) bool {
		calls++
		return true
	})
	assert.Equal(t, 2, calls)

	m.Reset()
	assert.Equal(t, 0, m.Count())
}

func TestContext_DoneChan(t *testing.T) {
	m := context.NewContext()
	id := m.Add(&context.Conversation{})

	ch := m.DoneChan(id)
	assert.NotNil(t, ch)

	m.Done(id)

	select {
	case <-ch:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("done chan not closed")
	}
}

func TestContext_ShutdownAll(t *testing.T) {
	m := context.NewContext()
	a := m.Add(&context.Conversation{})
	b := m.Add(&context.Conversation{})
	m.ShutdownAll()

	assert.NotNil(t, m.Get(a))
	assert.NotNil(t, m.Get(b))

	select {
	case <-m.DoneChan(a):
	case <-time.After(50 * time.Millisecond):
		t.Fatal("a not closed")
	}

	select {
	case <-m.DoneChan(b):
	case <-time.After(50 * time.Millisecond):
		t.Fatal("b not closed")
	}
}
