// File: mini/pkg/x_runtime/js/goja_test.go
package rtm_js

import (
	"context"
	"testing"

	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/stretchr/testify/require"
)

func TestJSRuntime_BasicExecution(t *testing.T) {
	rt := &JSRuntime{}
	err := rt.Init()
	require.NoError(t, err)

	action := act_core.NewAction("js.run", context.Background(), `1 + 2 + 3`)

	res, err := rt.Execute(action)
	require.NoError(t, err)
	require.Equal(t, int64(6), res)

	rt.Dispose()
	require.Nil(t, rt.vm)
}

func TestJSRuntime_NoCode(t *testing.T) {
	rt := &JSRuntime{}
	err := rt.Init()
	require.NoError(t, err)

	action := act_core.NewAction("js.empty", context.Background()) // no inputs

	res, err := rt.Execute(action)
	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, ErrNoScript, err)
}
