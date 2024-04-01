package logger

import "reflect"

// ifThen returns a if cond is true, otherwise returns b
func ifThen[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// ifThenFunc executes afn if cond is true, otherwise executes bfn
func ifThenFunc[T any](cond bool, afn func() T, bfn func() T) T {
	if cond {
		return afn()
	}
	return bfn()
}

// isArray returns true if v is a slice or array
func isArray(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
}

// lenIfArrayType returns the length of v if v is a slice or array
// otherwise returns false
func lenIfArrayType(v any) (length int, ok bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return 0, false
	}
	return rv.Len(), true
}
