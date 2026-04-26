package function

// If is a generic ternary operator. Returns ifVal when cond is true, otherwise elseVal.
func If[T any](cond bool, ifVal, elseVal T) T {
	if cond {
		return ifVal
	}
	return elseVal
}

// When applies fn to val if cond is true, otherwise returns val unchanged.
func When[T any](cond bool, val T, fn func(T) T) T {
	if cond {
		return fn(val)
	}
	return val
}

// IfElse lazily evaluates ifFn or elseFn based on cond.
// Useful when the branches are expensive to compute.
func IfElse[T any](cond bool, ifFn, elseFn func() T) T {
	if cond {
		return ifFn()
	}
	return elseFn()
}
