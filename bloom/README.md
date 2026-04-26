# bloom

概率型布隆过滤器。基于 Murmur3 哈希，可调节预期元素数量和误判率。仅依赖 `spaolacci/murmur3`。

## 快速开始

```go
import "github.com/tenz-io/gokit/bloom"

// 预期 10 万元素，误判率 0.01（1%）
bf := bloom.NewFilter(100_000, 0.01)
if bf == nil {
    panic("invalid parameters")
}

bf.Add([]byte("user-42"))
bf.AddString("user-99")

bf.Exists([]byte("user-42"))  // true
bf.ExistsString("user-99")    // true
bf.Exists([]byte("user-999")) // false (大概率)

fmt.Println(bf.ApproxCount())       // 2
fmt.Println(bf.FalsePositiveRate()) // ~0.0001 (当前估算)
```

## API 参考

### 构造

| 函数 | 签名 | 说明 |
|------|------|------|
| `NewFilter` | `func NewFilter(n uint64, p float64) Filter` | n 为预期元素数（必须 >0），p 为误判率（必须在 (0,1)）。参数无效返回 nil |

### Filter 接口

| 方法 | 签名 | 说明 |
|------|------|------|
| `Add` | `Add(data []byte)` | 插入元素 |
| `AddString` | `AddString(s string)` | 插入字符串元素（等价于 `Add([]byte(s))`） |
| `Exists` | `Exists(data []byte) bool` | 查询元素是否可能存在。false 表示一定不存在，true 表示可能存在（有误判率） |
| `ExistsString` | `ExistsString(s string) bool` | 字符串版查询 |
| `ApproxCount` | `ApproxCount() uint64` | 返回已插入的近似元素数（精确计数） |
| `FalsePositiveRate` | `FalsePositiveRate() float64` | 基于当前插入数量估算误判率。公式：`(1 - e^(-k*n/m))^k` |

## 参数配置

### 容量与精度权衡

| 场景 | n | p | 内存占用 | 哈希次数 |
|------|---|---|----------|----------|
| 小型缓存（1000 元素，1% 误判） | 1000 | 0.01 | ~1.2 KB | 7 |
| 中型去重（100 万元素，0.1%） | 1e6 | 0.001 | ~1.7 MB | 10 |
| 大型过滤（1 亿元素，0.01%） | 1e8 | 0.001 | ~167 MB | 10 |

计算方式：`m = -n * ln(p) / (ln(2))²`，`k = ln(2) * m / n`

### 运行时误判率

随着插入元素增加，实际误判率会上升。一旦 `ApproxCount` 接近或超过预期元素数 `n`，建议重建过滤器。

```go
if bf.FalsePositiveRate() > 0.05 {
    // 重建更大容量的过滤器
    bf = bloom.NewFilter(n*2, p)
}
```

## 最佳实践

### 选择合适的参数

- **误判率 p 不宜过高**：p=0.01 时，每 100 次不存在的查询会有 1 次误判，在安全敏感场景（如防止缓存穿透）应选用 p=0.001 或更低。
- **预期数量 n 应合理**：n 过小会导致过早饱和；n 过大浪费内存。建议设置为预期最大值的 1.2~1.5 倍。
- **始终检查 nil**：`NewFilter` 在参数无效时返回 nil，生产代码务必检查。

### 不适合的场景

- 需要删除元素 → 使用 cuckoo filter
- 需要精确计数 → 使用 hash set
- 数据量极小 → map 更简单直接

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| `NewFilter` 无效参数 | 未定义行为 | 返回 nil |
| `AddString` / `ExistsString` | 不存在 | 新增 |
| `ApproxCount` | 不存在 | 新增 |
| `FalsePositiveRate` | 不存在 | 新增 |
