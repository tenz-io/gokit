package logger

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
