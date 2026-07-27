package fn

// If 是通用三元运算符。cond 为 true 时返回 ifVal,否则返回 elseVal。
// 两个分支都会被急切求值;分支代价高时请用 IfElse。
func If[T any](cond bool, ifVal, elseVal T) T {
	if cond {
		return ifVal
	}
	return elseVal
}

// When 仅在 cond 为 true 时对 val 应用 fn,否则原样返回 val。
func When[T any](cond bool, val T, fn func(T) T) T {
	if cond {
		return fn(val)
	}
	return val
}

// IfElse 基于 cond 懒求值两个分支之一。当未选中的分支代价高或有副作用
// 时,用它(而非 If)。
func IfElse[T any](cond bool, ifFn, elseFn func() T) T {
	if cond {
		return ifFn()
	}
	return elseFn()
}

// Coalesce 返回首个不等于 T 零值的参数,若全部为零则返回零值。
// 它是 "首个非空" 习惯用法,适合用于挑选首个可用的配置值。
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

// Default 在 v 非零时返回 v,否则返回 def。它是 Coalesce 的双参特化。
func Default[T comparable](v, def T) T {
	var zero T
	if v != zero {
		return v
	}
	return def
}
