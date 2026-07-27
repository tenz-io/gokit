package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// lruManager 将泛型 [lruCache](K=string, V=[]byte)适配到
// [Manager] interface。值以 []byte 存储(blobs 用 JSON,Set/Get 用原始 bytes)。
// 没有默认 TTL:传入任何方法的非正 expire 表示"永不过期",与 local cache 一致。
type lruManager struct {
	c *lruCache[string, []byte]
}

// NewLRU 创建一个基于 LRU 容量限定的 cache,实现 [Manager]。
//
//   - capability 限定条目数;非正值默认为 120。
//   - onEvict 是一个可选回调,在条目因容量压力、过期或删除被 evict 时调用
//     (在 cache 锁外调用)。
//
// Options:[WithNow](可注入 clock)。[WithEvictInterval] 被忽略 ——
// LRU 在访问时 lazy 过期。
func NewLRU(
	capability int,
	onEvict func(key string, val []byte),
	opts ...Option,
) Manager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	if capability <= 0 {
		capability = 120
	}
	c := newLRU[string, []byte](capability, onEvict)
	c.withNow(o.nowFunc)
	return &lruManager{c: c}
}

func (m *lruManager) active() bool {
	return m != nil && m.c != nil
}

func (m *lruManager) Get(key string) (string, error) {
	if !m.active() {
		return "", ErrInactive
	}
	bs, ok := m.c.get(key)
	if !ok {
		return "", ErrNotFound
	}
	return string(bs), nil
}

func (m *lruManager) Set(key string, raw string, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	m.c.set(key, []byte(raw), expire)
	return nil
}

func (m *lruManager) SetNx(key string, raw string, expire time.Duration) (bool, error) {
	if !m.active() {
		return false, ErrInactive
	}
	// lruCache.setNx 在一把锁内完成存在性检查与写入,
	// 因此并发 SetNx 调用方不会同时看到"缺失"并同时写入。
	return m.c.setNx(key, []byte(raw), expire), nil
}

func (m *lruManager) GetBlob(key string, output any) error {
	if !m.active() {
		return ErrInactive
	}
	bs, ok := m.c.get(key)
	if !ok {
		return ErrNotFound
	}
	if err := decodeBlob(bs, output); err != nil {
		return err
	}
	return nil
}

func (m *lruManager) SetBlob(key string, val any, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	bs, err := encodeBlob(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	m.c.set(key, bs, expire)
	return nil
}

func (m *lruManager) Del(key string) error {
	if !m.active() {
		return ErrInactive
	}
	m.c.remove(key)
	return nil
}

func (m *lruManager) Expire(key string, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	// lruCache.expire 在一把锁内检查"存在且未过期",对缺失或已过期的 key
	// 返回 ok=false,因此我们既不会复活已过期条目,也不会与并发写者竞争。
	if ok := m.c.expire(key, expire); !ok {
		return ErrNotFound
	}
	return nil
}

// Close 对 LRU backend 是一个 no-op,因为它没有后台资源需要释放。
// 它满足 [Manager] 生命周期契约。
func (m *lruManager) Close() error { return nil }

// lruCache 是一个泛型、并发安全的 LRU cache,支持可选的每条目
// TTL 过期。它是 [lruManager] 的底层实现。零值 lruCache 不可直接使用;
// 请用 [newLRU] 构造。
//
// 每次变更与读取都由 sync.Mutex 保护,因此可跨 goroutine 共享
// 而无需外部加锁(不同于 v2 的 lru,后者明确不支持并发访问)。
//
// Eviction 回调(onEvict)在 cache 锁*外*运行:某方法在锁内
// 收集所有被清理的条目,释放锁,然后才调用回调。这让回调可以安全地
// 重入同一个 cache(例如读取或计数)而不死锁,也避免慢回调
// 阻塞无关的读者。
type lruCache[K comparable, V any] struct {
	capability int
	// onEvict 在非 nil 时,对每个被从 cache 清理的 item 调用:
	// 容量驱逐、显式 remove/removeOldest、过期以及 clear。
	// 它在锁释放后被调用(见类型注释)。
	onEvict func(key K, val V)

	ll      *list.List
	cache   map[K]*list.Element
	nowFunc func() time.Time

	mu sync.Mutex
}

// lruEntry 是 LRU 中的一个元素。它持有 key(用于 eviction 回调
// 与 map 删除)、value,以及一个绝对过期时间。
type lruEntry[K comparable, V any] struct {
	key      K
	val      V
	expireAt time.Time
}

// expired 判断条目 deadline 是否已过。零值 expireAt
// 表示该条目永不过期。边界是包含的:恰在 deadline 时刻
// 即视为已过期。
func (e *lruEntry[K, V]) expired(now time.Time) bool {
	if e.expireAt.IsZero() {
		return false
	}
	return !now.Before(e.expireAt)
}

// newLRU 创建一个 LRU cache。
//
//   - capability 限定条目数;非正值表示无限制
//     (此时过期是唯一的回收机制)。
//   - onEvict 是一个可选回调,在 item 被清理时调用(在锁外)。
//
// 没有默认 TTL:每个 set/setNx 都接受显式 duration,
// 非正值表示"永不过期"。该 cache 支持并发使用。
func newLRU[K comparable, V any](
	capability int,
	onEvict func(key K, val V),
) *lruCache[K, V] {
	return &lruCache[K, V]{
		capability: capability,
		ll:         list.New(),
		cache:      make(map[K]*list.Element),
		onEvict:    onEvict,
		nowFunc:    time.Now,
	}
}

// withNow 注入用于所有过期判断的 clock。返回 cache
// 以便链式调用。测试用它在不真实 sleep 的情况下推进时间。
func (c *lruCache[K, V]) withNow(now func() time.Time) *lruCache[K, V] {
	if now != nil {
		c.nowFunc = now
	}
	return c
}

// deadlineFor 根据每次调用的 duration 返回新条目的绝对过期时间。
// 非正 duration 表示条目永不过期(零值 time.Time);
// 正 duration 则返回 now+duration。
func deadlineFor(now time.Time, duration time.Duration) time.Time {
	if duration <= 0 {
		return time.Time{}
	}
	return now.Add(duration)
}

// set 以指定过期时间在 key 下新增(或更新)val。非正
// duration 表示条目永不过期。
func (c *lruCache[K, V]) set(key K, val V, duration time.Duration) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	evicted := c.setLocked(key, val, expireAt, now)
	c.mu.Unlock()

	c.fireOnEvict(evicted)
}

// setLocked 插入或更新 key,并返回因容量压力被清理的条目,
// 以便调用方在解锁后触发回调。调用方持有 c.mu。
func (c *lruCache[K, V]) setLocked(key K, val V, expireAt time.Time, now time.Time) []lruEntry[K, V] {
	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}
	var evicted []lruEntry[K, V]
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		e := ee.Value.(*lruEntry[K, V])
		e.val = val
		e.expireAt = expireAt
		return nil
	}
	ele := c.ll.PushFront(&lruEntry[K, V]{key: key, val: val, expireAt: expireAt})
	c.cache[key] = ele
	if c.capability > 0 && c.ll.Len() > c.capability {
		if e := c.removeOldestLocked(now); e != nil {
			evicted = append(evicted, *e)
		}
	}
	return evicted
}

// setNx 仅当 key 缺失或已过期时原子地把 val 写入 key。
// 当 key 存在且未过期时返回 existing=true(此时不发生写入),
// 写入成功时返回 existing=false。与"先 get 再 set"的序列不同,
// 存在性检查与写入在同一把锁内完成,因此并发 setNx 调用方
// 不会同时看到"缺失"并同时写入。由插入触发的容量驱逐
// 其 onEvict 回调在锁释放后触发。
func (c *lruCache[K, V]) setNx(key K, val V, duration time.Duration) (existing bool) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}
	var evicted []lruEntry[K, V]
	if ee, ok := c.cache[key]; ok {
		e := ee.Value.(*lruEntry[K, V])
		if !e.expired(now) {
			// 存在且未过期:不覆盖。
			c.mu.Unlock()
			return true
		}
		// 已过期:视作缺失。原地覆盖(无 eviction 回调
		// —— 与 set 的原地更新一致)。
		c.ll.MoveToFront(ee)
		e.val = val
		e.expireAt = expireAt
		c.mu.Unlock()
		return false
	}
	ele := c.ll.PushFront(&lruEntry[K, V]{key: key, val: val, expireAt: expireAt})
	c.cache[key] = ele
	if c.capability > 0 && c.ll.Len() > c.capability {
		if e := c.removeOldestLocked(now); e != nil {
			evicted = append(evicted, *e)
		}
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
	return false
}

// get 查找 key。命中时条目被移至链表前端(最近使用)。
// 已过期条目会被删除(其 onEvict 在解锁后触发)并
// 作为未命中返回。
func (c *lruCache[K, V]) get(key K) (val V, ok bool) {
	var zero V
	now := c.nowFunc()

	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return zero, false
	}
	ele, hit := c.cache[key]
	if !hit {
		c.mu.Unlock()
		return zero, false
	}
	if ele.Value.(*lruEntry[K, V]).expired(now) {
		evicted := []lruEntry[K, V]{*ele.Value.(*lruEntry[K, V])}
		c.removeElementLocked(ele)
		c.mu.Unlock()
		c.fireOnEvict(evicted)
		return zero, false
	}
	c.ll.MoveToFront(ele)
	result := ele.Value.(*lruEntry[K, V]).val
	c.mu.Unlock()
	return result, true
}

// expire 更新 key 的过期时间,并报告更新是否生效。非正
// duration 使条目永不过期。缺失或已过期的 key 不被改动,
// 并返回 ok=false —— expire 无法复活一个在逻辑上已过期的 key。
// 检查与更新在同一把锁内完成,因此不会有并发写者
// 插入两者之间。
func (c *lruCache[K, V]) expire(key K, duration time.Duration) (ok bool) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return false
	}
	ele, hit := c.cache[key]
	if !hit {
		return false
	}
	e := ele.Value.(*lruEntry[K, V])
	if e.expired(now) {
		// 不复活已过期条目;调用方应重新 set。
		return false
	}
	e.expireAt = expireAt
	c.ll.MoveToFront(ele)
	return true
}

// remove 从 cache 删除 key。缺失 key 为 no-op。
// eviction 回调(若有)在锁释放后触发。
func (c *lruCache[K, V]) remove(key K) {
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	var evicted []lruEntry[K, V]
	if ele, hit := c.cache[key]; hit {
		evicted = []lruEntry[K, V]{*ele.Value.(*lruEntry[K, V])}
		c.removeElementLocked(ele)
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// removeOldest 驱逐最久未使用的条目;eviction 回调
// 在解锁后触发。
func (c *lruCache[K, V]) removeOldest() {
	now := c.nowFunc()
	c.mu.Lock()
	e := c.removeOldestLocked(now)
	c.mu.Unlock()
	if e != nil {
		c.fireOnEvict([]lruEntry[K, V]{*e})
	}
}

// removeOldestLocked 驱逐链表尾部元素并返回其条目(不触发
// 回调)。调用方持有 c.mu。
func (c *lruCache[K, V]) removeOldestLocked(now time.Time) *lruEntry[K, V] {
	if c.cache == nil || c.ll == nil {
		return nil
	}
	ele := c.ll.Back()
	if ele == nil {
		return nil
	}
	e := ele.Value.(*lruEntry[K, V])
	c.removeElementLocked(ele)
	return e
}

// removeExpired 扫描并驱逐所有过期条目;回调在解锁后触发。
func (c *lruCache[K, V]) removeExpired() {
	now := c.nowFunc()
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	var evicted []lruEntry[K, V]
	for _, ele := range c.cache {
		e := ele.Value.(*lruEntry[K, V])
		if e.expired(now) {
			evicted = append(evicted, *e)
			c.removeElementLocked(ele)
		}
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// removeElementLocked 从 list 与 map 中移除 ele,但不触发
// 回调。调用方持有 c.mu。
func (c *lruCache[K, V]) removeElementLocked(ele *list.Element) {
	c.ll.Remove(ele)
	delete(c.cache, ele.Value.(*lruEntry[K, V]).key)
}

// len 返回当前条目数(包含尚未被 lazy 过期的条目)。
func (c *lruCache[K, V]) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// clear 删除所有条目,每个条目的 onEvict 在解锁后触发。
func (c *lruCache[K, V]) clear() {
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	evicted := make([]lruEntry[K, V], 0, len(c.cache))
	for _, ele := range c.cache {
		e := ele.Value.(*lruEntry[K, V])
		evicted = append(evicted, *e)
		c.ll.Remove(ele)
		delete(c.cache, e.key)
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// fireOnEvict 对每个条目调用 onEvict 回调。当不存在回调或
// 没有需要上报的内容时为 no-op。调用方在调用前必须已释放 c.mu,
// 因为回调可能重入 cache。
func (c *lruCache[K, V]) fireOnEvict(entries []lruEntry[K, V]) {
	if c.onEvict == nil || len(entries) == 0 {
		return
	}
	for _, e := range entries {
		c.onEvict(e.key, e.val)
	}
}
