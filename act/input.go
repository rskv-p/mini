package act

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
)

//---------------------
// Input Validation
//---------------------

// ValidateInputsNumber checks if the number of inputs is sufficient
func (action *Action) ValidateInputsNumber(length int) error {
	if len(action.Inputs) < length {
		log.Printf("Action %s requires at least %d inputs, but got %d", action.Name, length, len(action.Inputs))
		return fmt.Errorf("missing parameters: %s requires at least %d args", action.Name, length)
	}
	return nil
}

// NumberOfInputs returns the total number of inputs
func (action *Action) NumberOfInputs() int {
	return len(action.Inputs)
}

// NumberOfInputsIs checks if the number of inputs equals the specified number
func (action *Action) NumberOfInputsIs(num int) bool {
	return len(action.Inputs) == num
}

// InputNotNull ensures that a specific input is not nil
func (action *Action) InputNotNull(i int) error {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		log.Printf("Argument %d of %s cannot be nil", i, action.Name)
		return fmt.Errorf("argument %d of %s cannot be nil", i, action.Name)
	}
	return nil
}

//---------------------
// Common Parsers
//---------------------

// InputString parses the input as a string, using default values if necessary
func (action *Action) InputString(i int, defaults ...string) string {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return ""
	}
	if str, ok := action.Inputs[i].(string); ok {
		return str
	}
	return fmt.Sprintf("%v", action.Inputs[i])
}

// InputInt parses the input as an integer, using default values if necessary
func (action *Action) InputInt(i int, defaults ...int) int {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to 0", i, action.Name)
		return 0
	}

	switch v := action.Inputs[i].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("Error: failed to convert input at index %d to int: %v", i, err)
			return 0
		}
		return n
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected int, got %T", i, action.Name, v)
		return 0
	}
}

// InputUint32 parses the input as a uint32, using default values if necessary
func (action *Action) InputUint32(i int, defaults ...uint32) uint32 {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to 0", i, action.Name)
		return 0
	}

	switch v := action.Inputs[i].(type) {
	case uint32:
		return v
	case int:
		return uint32(v)
	case float64:
		return uint32(v)
	case string:
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			log.Printf("Error: failed to convert input at index %d to uint32: %v", i, err)
			return 0
		}
		return uint32(val)
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected uint32, got %T", i, action.Name, v)
		return 0
	}
}

// InputBool parses the input as a boolean, using default values if necessary
func (action *Action) InputBool(i int, defaults ...bool) bool {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to false", i, action.Name)
		return false
	}

	switch v := action.Inputs[i].(type) {
	case bool:
		return v
	case string:
		val, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Error: failed to convert input at index %d to bool: %v", i, err)
			return false
		}
		return val
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected bool, got %T", i, action.Name, v)
		return false
	}
}

//---------------------
// URL / Map Parsers
//---------------------

// InputURL parses the input as a URL and retrieves a specific key
func (action *Action) InputURL(i int, key string, defaults ...string) string {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to empty string", i, action.Name)
		return ""
	}

	switch v := action.Inputs[i].(type) {
	case url.Values:
		vals := v[key]
		if len(vals) > 0 {
			return vals[0]
		}
	case map[string]string:
		if val, ok := v[key]; ok {
			return val
		}
	case map[string]interface{}:
		if val, ok := v[key]; ok {
			return fmt.Sprintf("%v", val)
		}
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected url.Values, map[string]string, or map[string]interface{}, got %T", i, action.Name, v)
	}

	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

// InputMap parses the input as a map
func (action *Action) InputMap(i int) map[string]any {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to empty map", i, action.Name)
		return map[string]interface{}{}
	}

	switch v := action.Inputs[i].(type) {
	case map[string]interface{}:
		return v
	case url.Values:
		result := map[string]interface{}{}
		for key, vals := range v {
			if len(vals) == 1 {
				result[key] = vals[0]
			} else {
				result[key] = vals
			}
		}
		return result
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected map[string]interface{} or url.Values, got %T", i, action.Name, v)
	}

	return map[string]interface{}{}
}

//---------------------
// Slice Parsers
//---------------------

// InputArray parses the input as an array
func (action *Action) InputArray(i int) []any {
	if len(action.Inputs) <= i || action.Inputs[i] == nil {
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to empty array", i, action.Name)
		return nil
	}

	switch v := action.Inputs[i].(type) {
	case []interface{}:
		return v
	default:
		log.Printf("Error: input at index %d has an invalid type for action %s: expected []interface{}, got %T", i, action.Name, v)
		return nil
	}
}

// InputStrings parses the input as an array of strings
func (action *Action) InputStrings(i int) []string {
	action.ValidateInputsNumber(i + 1)

	switch values := action.Inputs[i].(type) {
	case []string:
		return values
	case []interface{}:
		strs := make([]string, 0, len(values))
		for _, val := range values {
			str, ok := val.(string)
			if !ok {
				log.Printf("Error: element in input at index %d is not a string, defaulting to empty array", i)
				return []string{}
			}
			strs = append(strs, str)
		}
		return strs
	default:
		log.Printf("Error: input at index %d is not an array of strings for action %s, defaulting to empty array", i, action.Name)
		return []string{}
	}
}

// InputsRecords parses the input as an array of records (maps)
func (action *Action) InputsRecords(i int) []map[string]any {
	action.ValidateInputsNumber(i + 1)
	raw := action.Inputs[i]
	if raw == nil {
		log.Printf("Warning: input at index %d is nil or missing for action %s, defaulting to empty array", i, action.Name)
		return []map[string]any{}
	}

	switch val := raw.(type) {
	case []map[string]interface{}:
		return val
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(val))
		for _, item := range val {
			switch v := item.(type) {
			case map[string]interface{}:
				out = append(out, v)
			default:
				log.Printf("Error: element in input at index %d is not a map, defaulting to empty array", i)
				return []map[string]any{}
			}
		}
		return out
	default:
		log.Printf("Error: input at index %d is not an array of maps for action %s, defaulting to empty array", i, action.Name)
		return []map[string]any{}
	}
}
