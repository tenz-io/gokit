// Package cache 提供纯进程内的 cache 原语:无界 map cache 和基于 capacity
// 限定的 LRU cache,二者均支持可选的每条目 TTL 过期。没有网络 backend,
// 没有 Redis,没有 Lua —— 一切都存活于本地进程内。
//
// V3 是对 cache/v2 的全新重写,移除了网络层并修正了并发契约:
//   - [Manager] interface 保留了 Get/Set/SetNx/GetBlob/SetBlob/Del/Expire,
//     并新增 Close 用于生命周期管理;但移除了 Eval(Lua 脚本需要远端存储),
//     并去掉了 context.Context 参数(纯内存 cache 无法响应取消,继续携带它是误导性的);
//   - [NewLocal] 返回一个基于 map 的 cache,带有一个后台清扫 goroutine
//     定期回收过期 key;Close 会停止它并等待其退出(v2 存在 goroutine 泄漏);
//   - [NewLRU] 返回一个基于 LRU 容量限定的 cache,底层为包内自带的
//     并发安全泛型 LRU;eviction 回调在锁外执行,
//     因此回调可以安全地重入 cache;
//   - 不设默认 TTL:每个 Set/SetNx/Expire 都接受显式 duration,
//     非正值表示"永不过期",正值则设置一个绝对 deadline。两个 backend 遵循
//     相同的契约,因此互换它们永远不会改变数据生命周期;
//   - 可注入的 clock([WithNow])让测试无需真实 sleep 即可驱动过期。
package cache

import (
	"errors"
	"time"
)

// Cache 操作错误。
var (
	// ErrNotFound 在 key 缺失或已过期时返回。
	ErrNotFound = errors.New("cache: key not found")
	// ErrInactive 在 cache 从未初始化(nil receiver)或已被关闭时返回。
	ErrInactive = errors.New("cache: inactive")
)

// Manager 是跨两种 backend 的统一 cache interface。业务代码
// 可在 [NewLocal] 与 [NewLRU] 之间互换而无需改动调用点;
// 两个实现遵循相同的过期契约。
//
// Expiration:每个 Set/SetNx/Expire 都接受显式的 expire duration。
// 非正 expire 表示 key 永不过期;正 expire 则设置一个相对于 now 的绝对
// deadline。这对两个 backend 都成立 —— 不存在会让切换 backend 的调用方
// 感到意外的默认 TTL。
//
// 读到的 key 的过期是 lazy 的:Get/GetBlob 对过期 key 返回 ErrNotFound 并
// 将其删除。对已过期 key 调用 Expire 是一个 no-op,并返回 ErrNotFound ——
// Expire 无法复活一个在逻辑上已过期的 key。
type Manager interface {
	// Get 返回 key 对应的原始字符串值;当 key 缺失或已过期时返回 ErrNotFound。
	// 已过期的 key 在读取时被 lazy 删除。
	Get(key string) (raw string, err error)
	// Set 将 raw 存入 key,并附带指定的过期时间。非正
	// expire 表示该 key 永不过期。
	Set(key string, raw string, expire time.Duration) (err error)
	// SetNx 仅当 key 不存在(或已过期)时将 raw 存入 key;
	// 当 key 已存在且未过期时返回 existing=true。存在性检查与写入
	// 相对于其他 SetNx/Set 调用是原子的。非正 expire 表示该 key 永不过期。
	SetNx(key string, raw string, expire time.Duration) (existing bool, err error)
	// GetBlob 取回 key 对应的 JSON 编码值并解码到 output(必须为指针)。
	// 缺失或过期时返回 ErrNotFound。JSON 解码发生在 cache 锁之外。
	GetBlob(key string, output any) (err error)
	// SetBlob 对 val 进行 JSON 编码并存入 key,附带指定的
	// 过期时间。非正 expire 表示该 key 永不过期。
	SetBlob(key string, val any, expire time.Duration) (err error)
	// Del 删除 key。当 key 缺失时为 no-op。
	Del(key string) (err error)
	// Expire 重置 key 的过期时间。非正 expire 使该 key 永不过期。
	// 当 key 缺失或已过期时返回 ErrNotFound(它不会复活已过期的 key)。
	Expire(key string, expire time.Duration) (err error)
	// Close 释放任何后台资源(例如 [NewLocal] 中的清扫 goroutine)。
	// 它是幂等的,可安全多次调用;没有资源的 backend(LRU)将其视为 no-op。
	// Close 之后 cache 仍可用于读/写;仅后台回收停止,
	// 且 Close 返回时清扫 goroutine 保证已退出。
	Close() error
}

// Option 在构造期配置一个 cache。它由 [NewLocal] 与 [NewLRU] 共享;
// 与某 backend 无关的 option 会被该 backend 忽略。
type Option func(*options)

type options struct {
	nowFunc       func() time.Time
	evictInterval time.Duration
}

func defaultOptions() options {
	return options{
		nowFunc:       time.Now,
		evictInterval: 5 * time.Minute,
	}
}

// WithNow 注入用于所有过期判断的 clock。测试传入可控时间源,
// 从而无需真实 sleep 即可推进过期;生产环境保持为 time.Now。
func WithNow(now func() time.Time) Option {
	return func(o *options) {
		if now != nil {
			o.nowFunc = now
		}
	}
}

// WithEvictInterval 设置 local cache 后台清扫的运行间隔。
// 它仅作用于 [NewLocal];[NewLRU] 在访问时 lazy 过期。非正值
// 完全禁用后台清扫(过期仍在读取时 lazy 生效)。
func WithEvictInterval(d time.Duration) Option {
	return func(o *options) {
		o.evictInterval = d
	}
}
