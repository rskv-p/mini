// file:mini/mod/m_api/api_test.go
package api_mod_test

import (
	"errors"
	"testing"

	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_api/api_mod"

	"github.com/stretchr/testify/require"
)

func TestRegisterAction_Public(t *testing.T) {
	//---------------------
	// Register public handler
	//---------------------
	api_mod.RegisterAction("test.echo", func(a act_type.IAction) any {
		return "ok"
	}, api_mod.APIOption{Public: true})

	entry, ok := api_mod.Get("test.echo")
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
	api_mod.RegisterAction("test.hidden", func(a act_type.IAction) any {
		return "secret"
	}, api_mod.APIOption{Public: false})

	_, ok := api_mod.Get("test.hidden")
	require.False(t, ok) // not exposed via API
}

func TestRegisterAction_ErrorPropagation(t *testing.T) {
	//---------------------
	// Register that returns error
	//---------------------
	api_mod.RegisterAction("test.fail", func(a act_type.IAction) any {
		panic(errors.New("boom"))
	}, api_mod.APIOption{Public: true})

	entry, ok := api_mod.Get("test.fail")
	require.True(t, ok)

	_, err := entry.Handler([]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
}
