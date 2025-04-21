// file:mini/pkg/x_req/request_test.go
package bus_req_test

import (
	"testing"

	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/stretchr/testify/require"
)

//---------------------
// Basic
//---------------------

func TestRequest_Basic(t *testing.T) {
	req := bus_req.NewTestRequest("demo.test", []byte("ping"))

	// Test headers
	req.SetHeader("Authorization", "Bearer abc")
	require.Equal(t, "Bearer abc", req.Headers().Get("Authorization"))

	// Test RespondJSON
	var responded []byte
	req.Respond = func(data []byte) error {
		responded = data
		return nil
	}
	err := req.RespondJSON(map[string]string{"msg": "ok"})
	require.NoError(t, err)
	require.JSONEq(t, `{"msg":"ok"}`, string(responded))

	// Test Error()
	var gotErr string
	req.Err = func(code, msg string, _ any) error {
		gotErr = code + ":" + msg
		return nil
	}
	_ = req.Error("403", "forbidden", nil)
	require.Equal(t, "403:forbidden", gotErr)
}
