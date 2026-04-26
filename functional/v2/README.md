# functional

泛型函数式编程：Map、Filter、Reduce、TopK 等。

```go
import "github.com/tenz-io/gokit/functional/v2"
```

## 快速开始

```go
ids := function.Map(users, func(u User) int { return u.ID })
active := function.Filter(users, func(u User) bool { return u.Active })
sum := function.Reduce(nums, func(acc, n int) int { return acc + n }, 0)
top3 := function.TopK(items, 3, func(a, b Item) bool { return a.Score < b.Score })
```
