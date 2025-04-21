// file:mini/mod/m_api/api_test.go
package m_api_test

import (
	"errors"
	"testing"

	"github.com/rskv-p/mini/mod/m_api"
	"github.com/rskv-p/mini/typ"

	"github.com/stretchr/testify/require"
)

func TestRegisterAction_Public(t *testing.T) {
	//---------------------
	// Register public handler
	//---------------------
	m_api.RegisterAction("test.echo", func(a typ.IAction) any {
		return "ok"
	}, m_api.APIOption{Public: true})

	entry, ok := m_api.Get("test.echo")
	require.True(t, ok)
	require.True(t, entry.Options.Public)

	//---------------------
	// Call wrapped handler
	//---------------------
	res, err := entry.Handler([]any{})
	require.NoError(t, err)
	require.Equal(t, "ok", res)
}

func TestRegisterAction_Private(t *testing.T) {
	//---------------------
	// Register private handler
	//---------------------
	m_api.RegisterAction("test.hidden", func(a typ.IAction) any {
		return "secret"
	}, m_api.APIOption{Public: false})

	_, ok := m_api.Get("test.hidden")
	require.False(t, ok) // not exposed via API
}

func TestRegisterAction_ErrorPropagation(t *testing.T) {
	//---------------------
	// Register that returns error
	//---------------------
	m_api.RegisterAction("test.fail", func(a typ.IAction) any {
		panic(errors.New("boom"))
	}, m_api.APIOption{Public: true})

	entry, ok := m_api.Get("test.fail")
	require.True(t, ok)

	_, err := entry.Handler([]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
}
