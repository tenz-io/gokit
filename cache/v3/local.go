package cache

import (
	"fmt"
	"sync"
	"time"
)

// item 是单个 local-cache 条目。raw 是存储的 bytes;expireAt
// 是绝对 deadline,零值表示"永不过期"。使用 time.Time(而非 unix 秒整数)
// 以保持亚秒级 TTL 的精度。
type item struct {
	raw      []byte
	expireAt time.Time
}

// localCache 是一个进程本地的 map cache,支持可选的每条目 TTL,
// 并带有一个后台清扫 goroutine,定期回收过期 key,避免在读稀疏时内存无限增长。
//
// 它实现了 [Manager]。通过 [NewLocal] 构造。
//
// Concurrency:每次变更与读取都由 RWMutex 保护。读取时的 lazy 过期
// 是*同步且带条件的*:一个发现已过期 key 的 Get 会释放读锁,
// 获取写锁,重新确认同一 item 仍然存在且仍过期,然后才删除它。
// 这避免了如下竞争:异步删除恰好把并发 Set 刚写入同 key 的值删掉。
// GetBlob 额外在锁内复制原始 bytes,解锁后再解码 JSON,
// 因此一个缓慢或重入的 Unmarshal 不会阻塞写者或死锁。
type localCache struct {
	m       map[string]*item
	nowFunc func() time.Time
	lock    sync.RWMutex

	// sweepWg 跟踪后台清扫 goroutine,以便 Close 能等待其退出。
	// stopCh 在非 nil 时被关闭以通知清扫停止。
	stopCh  chan struct{}
	sweepWg sync.WaitGroup
	startMu sync.Mutex
	started bool
}

// NewLocal 创建一个进程本地的 map cache。默认情况下它会启动一个后台
// goroutine,每 5 分钟清扫一次过期 key;传入
// [WithEvictInterval](0) 可禁用它(过期仍在读取时 lazy 生效)。
// 返回的 cache 在使用完毕后必须 Close 以停止清扫。
//
// Options:[WithNow](可注入 clock),[WithEvictInterval]。
func NewLocal(opts ...Option) Manager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	lc := &localCache{
		m:       make(map[string]*item),
		nowFunc: o.nowFunc,
	}
	lc.startEvict(o.evictInterval)
	return lc
}

// Close 停止后台清扫 goroutine 并阻塞,直到它已退出。
// 可多次调用,也可对一个禁用了清扫的 cache 调用,均安全。
// Close 之后 cache 内容仍可用;仅回收循环停止。
func (lc *localCache) Close() error {
	if lc == nil {
		return nil
	}
	lc.startMu.Lock()
	if lc.started && lc.stopCh != nil {
		select {
		case <-lc.stopCh:
			// 已关闭
		default:
			close(lc.stopCh)
		}
		lc.started = false
	}
	lc.startMu.Unlock()
	// 在 startMu 外等待,这样清扫 goroutine(它在退出时不触碰
	// startMu)才能真正终止。
	lc.sweepWg.Wait()
	return nil
}

func (lc *localCache) active() bool {
	return lc != nil && lc.m != nil
}

// startEvict 按指定间隔启动后台清扫。非正间隔
// 会禁用清扫。
func (lc *localCache) startEvict(interval time.Duration) {
	if !lc.active() || interval <= 0 {
		return
	}
	lc.startMu.Lock()
	defer lc.startMu.Unlock()
	if lc.started {
		return
	}
	lc.stopCh = make(chan struct{})
	lc.started = true
	lc.sweepWg.Add(1)
	go lc.evictLoop(interval)
}

func (lc *localCache) evictLoop(interval time.Duration) {
	defer lc.sweepWg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-lc.stopCh:
			return
		case <-ticker.C:
			lc.evict()
		}
	}
}

// evict 在写锁下删除所有过期条目。
func (lc *localCache) evict() {
	if !lc.active() {
		return
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()

	now := lc.nowFunc()
	for k, v := range lc.m {
		if !v.expireAt.IsZero() && !now.Before(v.expireAt) {
			delete(lc.m, k)
		}
	}
}

// expireAt 返回新条目的绝对 deadline。非正 duration
// 表示"永不过期"(零值 time.Time)。
func (lc *localCache) expireAt(expire time.Duration) time.Time {
	if expire <= 0 {
		return time.Time{}
	}
	return lc.nowFunc().Add(expire)
}

// expired 判断是否已超过条目的 deadline。零值 expireAt
// 表示该条目永不过期。
func (lc *localCache) expired(it *item) bool {
	if it.expireAt.IsZero() {
		return false
	}
	return !lc.nowFunc().Before(it.expireAt)
}

// deleteIfExpired 仅当 cache 仍持有完全相同的 item(按指针比较)
// 且该 item 仍过期时才删除 key。调用方刚在读锁下观察到该 item;
// 在写锁下再次确认可关闭如下时间窗口:并发的 Set 可能已替换了该值 ——
// 若确实如此,我们不动新值。调用时读锁必须已释放。
func (lc *localCache) deleteIfExpired(key string, stale *item) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	cur, ok := lc.m[key]
	if !ok || cur != stale {
		// 已被并发替换或删除 —— 不动新值。
		return
	}
	if !lc.expired(cur) {
		// 时间被回拨或被重新 Expire 为有效 deadline。
		return
	}
	delete(lc.m, key)
}

func (lc *localCache) Get(key string) (raw string, err error) {
	if !lc.active() {
		return "", ErrInactive
	}

	lc.lock.RLock()
	it, found := lc.m[key]
	if !found || it == nil {
		lc.lock.RUnlock()
		return "", ErrNotFound
	}
	expired := lc.expired(it)
	val := string(it.raw)
	lc.lock.RUnlock()

	if expired {
		lc.deleteIfExpired(key, it)
		return "", ErrNotFound
	}
	return val, nil
}

func (lc *localCache) Set(key string, raw string, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) SetNx(key string, raw string, expire time.Duration) (existing bool, err error) {
	if !lc.active() {
		return false, ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if it, ok := lc.m[key]; ok && it != nil {
		// 将已过期 key 视作缺失,以便 SetNx 写入。
		if !lc.expired(it) {
			return true, nil
		}
	}
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return false, nil
}

func (lc *localCache) GetBlob(key string, output any) error {
	if !lc.active() {
		return ErrInactive
	}

	lc.lock.RLock()
	it, found := lc.m[key]
	if !found || it == nil {
		lc.lock.RUnlock()
		return ErrNotFound
	}
	expired := lc.expired(it)
	// 在锁内复制原始 bytes;解锁后再解码,这样缓慢或重入的
	// Unmarshal 不会阻塞写者或死锁。
	raw := make([]byte, len(it.raw))
	copy(raw, it.raw)
	lc.lock.RUnlock()

	if expired {
		lc.deleteIfExpired(key, it)
		return ErrNotFound
	}
	if err := decodeBlob(raw, output); err != nil {
		return err
	}
	return nil
}

func (lc *localCache) SetBlob(key string, val any, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	bs, err := encodeBlob(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: bs, expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) Del(key string) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	delete(lc.m, key)
	return nil
}

func (lc *localCache) Expire(key string, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	it, ok := lc.m[key]
	if !ok || it == nil {
		return ErrNotFound
	}
	// 不复活已过期的 key。
	if lc.expired(it) {
		return ErrNotFound
	}
	it.expireAt = lc.expireAt(expire)
	return nil
}
