package codec_test

import (
	"testing"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	data, err := codec.Marshal(map[string]string{"foo": "bar"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"foo":"bar"}`, string(data))
}

func TestUnmarshal(t *testing.T) {
	jsonStr := []byte(`{"key":123}`)
	var out map[string]any
	err := codec.Unmarshal(jsonStr, &out)
	assert.NoError(t, err)
	assert.Equal(t, float64(123), out["key"])
}

func TestUnmarshalInvalid(t *testing.T) {
	var out map[string]any
	err := codec.Unmarshal([]byte(`{invalid`), &out)
	assert.Error(t, err)
}

func TestMustMarshalSuccess(t *testing.T) {
	data := codec.MustMarshal(map[string]string{"x": "y"})
	assert.JSONEq(t, `{"x":"y"}`, string(data))
}

func TestMustMarshalPanic(t *testing.T) {
	// тип, который не может быть сериализован (chan)
	type Bad struct {
		C chan int
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on MustMarshal, got none")
		}
	}()

	_ = codec.MustMarshal(Bad{}) // должен паниковать
}
