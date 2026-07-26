package fn

// If is a generic ternary operator. Returns ifVal when cond is true, else
// elseVal. Both branches are eagerly evaluated; for expensive branches use
// IfElse.
func If[T any](cond bool, ifVal, elseVal T) T {
	if cond {
		return ifVal
	}
	return elseVal
}

// When applies fn to val only when cond is true, otherwise returns val
// unchanged.
func When[T any](cond bool, val T, fn func(T) T) T {
	if cond {
		return fn(val)
	}
	return val
}

// IfElse lazily evaluates one of two branches based on cond. Use it (over If)
// when the unselected branch would be expensive or have side effects to
// compute.
func IfElse[T any](cond bool, ifFn, elseFn func() T) T {
	if cond {
		return ifFn()
	}
	return elseFn()
}

// Coalesce returns the first argument that is not the zero value of T, or the
// zero value if all are zero. It is the "first non-empty" idiom, useful for
// picking the first available config value.
//
//	Coalesce(envValue, flagValue, defaultValue)
func Coalesce[T comparable](vs ...T) T {
	var zero T
	for _, v := range vs {
		if v != zero {
			return v
		}
	}
	return zero
}

// Default returns v if it is non-zero, otherwise def. It is the two-argument
// specialization of Coalesce.
func Default[T comparable](v, def T) T {
	var zero T
	if v != zero {
		return v
	}
	return def
}
