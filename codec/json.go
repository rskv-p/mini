// file: mini/codec/json.go
package codec

import (
	"encoding/json"
	"fmt"
)

// Marshal is a helper for encoding any value to JSON.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal is a helper for decoding JSON into a target value.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MustMarshal marshals a value or panics on failure.
func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal error: %v", err))
	}
	return b
}
