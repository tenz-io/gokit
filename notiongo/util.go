package notiongo

import "reflect"

func ifThen[T any](cond bool, tt, tf T) T {
	if cond {
		return tt
	}
	return tf
}

func ifThenFn[T any](cond bool, tt, tf func() T) T {
	if cond {
		return tt()
	}
	return tf()
}

type ValT interface {
	string | int | float64 | bool | map[string]any | []any
}

// ValueOrDefault returns the value if it is not empty, otherwise it returns the default value.
// Works for string, int, float64, bool, map[string]any, []any, and nil
func ValueOrDefault[ValT any](value any, defaultVal ValT) ValT {
	// If value is nil, immediately return the default value
	if value == nil {
		return defaultVal
	}

	// Get the reflect.Type of T (the default value's type)
	defaultValueType := reflect.TypeOf(defaultVal)

	// Handle based on the type of defaultVal
	switch defaultValueType.Kind() {
	case reflect.String:
		if v, ok := value.(string); ok && v != "" {
			return any(v).(ValT)
		}
	case reflect.Int:
		if v, ok := value.(int); ok && v != 0 {
			return any(v).(ValT)
		}
	case reflect.Float64:
		if v, ok := value.(float64); ok && v != 0.0 {
			return any(v).(ValT)
		}
	case reflect.Bool:
		if v, ok := value.(bool); ok {
			return any(v).(ValT)
		}
	case reflect.Map:
		if v, ok := value.(map[string]any); ok && len(v) > 0 {
			return any(v).(ValT)
		}
	case reflect.Slice:
		if v, ok := value.([]any); ok && len(v) > 0 {
			return any(v).(ValT)
		}
	default:
	}

	// If value does not match the expected type or is the zero value, return defaultVal
	return defaultVal
}
