// file: mini/codec/helper_test.go
package codec_test

import (
	"testing"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	msg := codec.NewMessage("string")
	msg.Set("str", "abc")
	msg.Set("int", 123)
	msg.Set("bytes", []byte("xyz"))
	msg.Set("float", 3.5)
	msg.Set("bool", true)
	msg.Set("obj", map[string]any{"x": 1})

	assert.Equal(t, "abc", msg.GetString("str"))
	assert.Equal(t, "123", msg.GetString("int"))
	assert.Equal(t, "xyz", msg.GetString("bytes"))
	assert.Equal(t, "3.5", msg.GetString("float"))
	assert.Equal(t, "true", msg.GetString("bool"))
	assert.Contains(t, msg.GetString("obj"), `"x":1`)
}

func TestGetInt(t *testing.T) {
	msg := codec.NewMessage("int")
	msg.Set("int", 42)
	msg.Set("int64", int64(99))
	msg.Set("float", 1.9)
	msg.Set("str", "100")
	msg.Set("bad", "xx")

	assert.Equal(t, int64(42), msg.GetInt("int"))
	assert.Equal(t, int64(99), msg.GetInt("int64"))
	assert.Equal(t, int64(1), msg.GetInt("float"))
	assert.Equal(t, int64(100), msg.GetInt("str"))
	assert.Equal(t, int64(0), msg.GetInt("bad"))
}

func TestGetFloat(t *testing.T) {
	msg := codec.NewMessage("float")
	msg.Set("f", 1.23)
	msg.Set("i", 5)
	msg.Set("i64", int64(8))
	msg.Set("str", "3.14")
	msg.Set("bad", "x")

	assert.InDelta(t, 1.23, msg.GetFloat("f"), 0.001)
	assert.InDelta(t, 5.0, msg.GetFloat("i"), 0.001)
	assert.InDelta(t, 8.0, msg.GetFloat("i64"), 0.001)
	assert.InDelta(t, 3.14, msg.GetFloat("str"), 0.001)
	assert.Equal(t, float64(0), msg.GetFloat("bad"))
}

func TestGetBool(t *testing.T) {
	msg := codec.NewMessage("bool")
	msg.Set("b", true)
	msg.Set("str1", "true")
	msg.Set("str0", "false")
	msg.Set("bad", "xx")

	assert.Equal(t, true, msg.GetBool("b"))
	assert.Equal(t, true, msg.GetBool("str1"))
	assert.Equal(t, false, msg.GetBool("str0"))
	assert.Equal(t, false, msg.GetBool("bad"))
}
