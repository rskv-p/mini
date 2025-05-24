package codec

import (
	"encoding/json"
	"strconv"
)

// GetString returns the string value of key in body.
func (m *Message) GetString(key string) string {
	v, _ := m.Get(key)
	s, _ := toString(v)
	return s
}

// GetInt returns the int64 value of key in body.
func (m *Message) GetInt(key string) int64 {
	v, _ := m.Get(key)
	i, _ := toInt64(v)
	return i
}

// GetFloat returns the float64 value of key in body.
func (m *Message) GetFloat(key string) float64 {
	v, _ := m.Get(key)
	f, _ := toFloat64(v)
	return f
}

// GetBool returns the boolean value of key in body.
func (m *Message) GetBool(key string) bool {
	v, _ := m.Get(key)
	b, _ := toBool(v)
	return b
}

// toString tries to convert any value to a string.
func toString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case []byte:
		return string(x), true
	case int:
		return strconv.Itoa(x), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	default:
		b, err := json.Marshal(x)
		return string(b), err == nil
	}
}

// toInt64 tries to convert any value to int64.
func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		return i, err == nil
	}
	return 0, false
}

// toFloat64 tries to convert any value to float64.
func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

// toBool tries to convert any value to bool.
func toBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		b, err := strconv.ParseBool(x)
		return b, err == nil
	}
	return false, false
}
