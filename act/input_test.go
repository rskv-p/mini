// file:mini/act/input_test.go
package act_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/rskv-p/mini/act"

	"github.com/stretchr/testify/require"
)

func TestAction_Parsers(t *testing.T) {
	// Create a context (use context.Background() or context.TODO())
	ctx := context.Background()

	// Create a new action with the context and the input arguments
	a := act.NewAction("test.echo", ctx,
		"hello", 123, "42", true,
		map[string]any{"key": "value"},
		url.Values{"foo": {"bar"}},
		[]interface{}{"a", "b"},
		[]map[string]any{{"id": 1}, {"id": 2}},
	)

	// String
	require.Equal(t, "hello", a.InputString(0))
	require.Equal(t, "fallback", a.InputString(99, "fallback"))

	// Int
	require.Equal(t, 123, a.InputInt(1))
	require.Equal(t, 42, a.InputInt(2))
	require.Equal(t, 999, a.InputInt(99, 999))

	// Bool
	require.Equal(t, true, a.InputBool(3))
	require.Equal(t, false, a.InputBool(99, false))

	// Map
	m := a.InputMap(4)
	require.Equal(t, "value", m["key"])

	// URL field
	require.Equal(t, "bar", a.InputURL(5, "foo"))
	require.Equal(t, "none", a.InputURL(5, "missing", "none"))

	// Array
	arr := a.InputArray(6)
	require.Len(t, arr, 2)
	require.Equal(t, "a", arr[0])

	// Strings
	strs := a.InputStrings(6)
	require.Equal(t, []string{"a", "b"}, strs)

	// Records
	records := a.InputsRecords(7)
	require.Len(t, records, 2)
	require.Equal(t, 1, records[0]["id"])
}
