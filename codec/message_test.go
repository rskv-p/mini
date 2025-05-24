package codec_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	msg := codec.NewMessage("test")
	assert.Equal(t, "test", msg.GetType())
	assert.NotNil(t, msg.GetBodyMap())
	assert.NotNil(t, msg.GetHeaders())
}

func TestNewRequest(t *testing.T) {
	msg := codec.NewRequest("service.node", "ctx123")
	assert.Equal(t, "request", msg.GetType())
	assert.Equal(t, "service.node", msg.GetNode())
	assert.Equal(t, "ctx123", msg.GetContextID())
}

func TestNewResponse(t *testing.T) {
	msg := codec.NewResponse("ctx789", 200)
	assert.Equal(t, "response", msg.GetType())
	assert.Equal(t, "ctx789", msg.GetContextID())
	assert.Equal(t, 200, msg.StatusCode)
}

func TestSetAndGet(t *testing.T) {
	msg := codec.NewMessage("data")
	msg.Set("foo", "bar")
	val, ok := msg.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", msg.GetString("foo"))
}

func TestGetTypedValues(t *testing.T) {
	msg := codec.NewMessage("typed")
	msg.Set("int", 42)
	msg.Set("float", 3.14)
	msg.Set("bool", true)

	assert.Equal(t, int64(42), msg.GetInt("int"))
	assert.InDelta(t, 3.14, msg.GetFloat("float"), 0.001)
	assert.Equal(t, true, msg.GetBool("bool"))
}

func TestHeaders(t *testing.T) {
	msg := codec.NewMessage("header")
	msg.SetHeader("X-Test", "abc")
	assert.Equal(t, "abc", msg.GetHeader("X-Test"))
	assert.Equal(t, "", msg.GetHeader("Missing"))
}

func TestSetErrorAndResult(t *testing.T) {
	msg := codec.NewMessage("err")

	msg.SetError(errors.New("failed"))
	assert.True(t, msg.HasError())
	assert.Equal(t, "failed", msg.GetError())

	msg.SetError(nil)
	assert.False(t, msg.HasError())

	msg.SetResult(map[string]string{"key": "value"})

	var res map[string]string
	err := msg.GetResult(&res)
	assert.NoError(t, err)
	assert.Equal(t, "value", res["key"])
}

func TestSetBodyAndRawBody(t *testing.T) {
	msg := codec.NewMessage("raw")
	msg.SetBody(map[string]any{"x": 1, "y": true})
	raw := msg.GetRawBody()
	assert.NotNil(t, raw)
	assert.Contains(t, string(raw), `"x":1`)
}

func TestSetBodyNil(t *testing.T) {
	msg := codec.NewMessage("nil")
	msg.SetBody(nil)
	assert.Nil(t, msg.Body)
	assert.Nil(t, msg.GetRawBody()) // Should not panic
}

func TestGetResultMissing(t *testing.T) {
	msg := codec.NewMessage("empty")
	var out any
	err := msg.GetResult(&out)
	assert.NoError(t, err)
}

func TestHasErrorNonString(t *testing.T) {
	msg := codec.NewMessage("errtype")
	msg.Set("error", 123)
	assert.True(t, msg.HasError()) // will still stringify
}

func TestCopyMessage(t *testing.T) {
	orig := codec.NewMessage("copy")
	orig.Set("key", "val")
	orig.SetHeader("X", "1")
	_ = orig.UpdateRawBody()

	copy := orig.Copy().(*codec.Message)

	assert.Equal(t, "val", copy.GetString("key"))
	assert.Equal(t, "1", copy.GetHeader("X"))

	assert.False(t, &orig.Body == &copy.Body)
	assert.False(t, &orig.Headers == &copy.Headers)

	assert.Equal(t, orig.Type, copy.Type)
	assert.Equal(t, orig.RawBody, copy.RawBody)
}

func TestNilCopy(t *testing.T) {
	var nilMsg *codec.Message
	assert.Nil(t, nilMsg.Copy())
}

func TestValidate(t *testing.T) {
	valid := codec.NewRequest("service", "ctx")
	assert.NoError(t, valid.Validate())

	invalid := codec.NewMessage("")
	err := invalid.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Type")

	req := codec.NewMessage("request")
	req.SetContextID("123")
	err = req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Node")

	noCtx := codec.NewMessage("ping")
	err = noCtx.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ContextID")
}

func TestUpdateRawBody(t *testing.T) {
	msg := codec.NewMessage("rawtest")
	msg.Set("a", 1)
	err := msg.UpdateRawBody()
	assert.NoError(t, err)

	var out map[string]any
	err = json.Unmarshal(msg.GetRawBody(), &out)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), out["a"])
}
