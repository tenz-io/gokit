package app

import (
	"fmt"
)

// 本文件提供类型安全的适配器,消除调用方在每个 Init/Run 回调中重复的
// 类型断言样板。Config.Conf 是 any,因为 Inits 切片是异构的(每个 init
// 看到的可能不是同一类型),但单个具体回调几乎总是固定到调用方的
// Config 类型;这些适配器把那一次断言集中到此处,失败即转成 error 走
// 既有的失败路径,调用方得到强类型 conf。

// TypedInitFunc 与 InitFunc 等价,但 conf 已断言为 *T,无需调用方手写
// "conf.(*MyConfig), ok := ..."。返回值仍是 CleanFunc + error,与框架一致。
type TypedInitFunc[T any] func(c *Context, conf *T) (CleanFunc, error)

// AdaptInit 把 TypedInitFunc 包装成框架需要的 InitFunc。conf 非 *T 时
// 返回 error(Init 失败),而不是 panic。
func AdaptInit[T any](fn TypedInitFunc[T]) InitFunc {
	return func(c *Context, conf any) (CleanFunc, error) {
		typed, ok := conf.(*T)
		if !ok {
			return nil, fmt.Errorf("app: init: conf type %T is not *%T", conf, new(T))
		}
		return fn(c, typed)
	}
}

// TypedRunFunc 与 RunFunc 等价,但 conf 已断言为 *T;返回 error 而非
// 通过 errC 报告:返回 nil 视为干净完成并触发关闭,非 nil 视为致命
// 错误(等价于旧 errC <- err)。
type TypedRunFunc[T any] func(c *Context, conf *T) error

// AdaptRun 把 TypedRunFunc 包装成框架需要的 RunFunc。它接管 errC:
// fn 返回非 nil error 时发往 errC(致命),返回 nil 时发 nil(干净完成)。
// 典型用法见 example。
func AdaptRun[T any](fn TypedRunFunc[T]) RunFunc {
	return func(c *Context, conf any, errC chan<- error) {
		typed, ok := conf.(*T)
		if !ok {
			errC <- fmt.Errorf("app: run: conf type %T is not *%T", conf, new(T))
			return
		}
		errC <- fn(c, typed)
	}
}
