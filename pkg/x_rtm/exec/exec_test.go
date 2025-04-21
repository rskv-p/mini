// file:mini/pkg/x_rtm/exec/exec_test.go
package exec_test

import (
	"context"
	"testing"
	"time"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/pkg/x_rtm/exec"

	"github.com/stretchr/testify/require"
)

func TestExecRuntime_Echo(t *testing.T) {
	rt := exec.New()
	require.NoError(t, rt.Init())

	action := act.NewAction("echo.test", context.Background(), "echo", "Hello", "World").
		WithContext(context.Background())

	res, err := rt.Execute(action.(*act.Action))
	require.NoError(t, err)

	out := res.(map[string]any)
	require.Contains(t, out["stdout"], "Hello World")
	require.NotZero(t, out["pid"])
}

func TestExecRuntime_Timeout(t *testing.T) {
	rt := exec.New()
	require.NoError(t, rt.Init())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	action := act.NewAction("sleep.test", context.Background(), "sleep", "2").
		WithContext(ctx)

	res, err := rt.Execute(action.(*act.Action))
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "signal: killed")
}

func TestExecRuntime_Stop(t *testing.T) {
	rt := exec.New()
	require.NoError(t, rt.Init())

	action := act.NewAction("long.running", context.Background(), "sleep", "5")

	go func() {
		_, _ = rt.Execute(action)
	}()

	time.Sleep(300 * time.Millisecond)

	list := rt.List()
	require.Contains(t, list, "long.running")

	err := rt.Stop("long.running")
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	list = rt.List()
	require.NotContains(t, list, "long.running")
}

func TestExecRuntime_Stderr(t *testing.T) {
	rt := exec.New()
	require.NoError(t, rt.Init())

	action := act.NewAction("invalid.cmd", context.Background(), "ls", "/this/should/not/exist")

	res, err := rt.Execute(action)
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "No such file or directory")
}

func TestExecRuntime_StopProcess(t *testing.T) {
	rt := exec.New()
	require.NoError(t, rt.Init())

	action := act.NewAction("stop.test", context.Background(), "sleep", "5")

	go func() {
		_, _ = rt.Execute(action)
	}()

	time.Sleep(300 * time.Millisecond)

	list := rt.List()
	require.Contains(t, list, "stop.test")

	err := rt.Stop("stop.test")
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	list = rt.List()
	require.NotContains(t, list, "stop.test")
}
