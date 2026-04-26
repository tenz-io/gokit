# collection

泛型数据结构：Stack、Queue、PriorityQueue、Set。

```go
import "github.com/tenz-io/gokit/collection/v2"
```

## 快速开始

```go
s := collection.NewStack[int](); s.Push(1); v, _ := s.Pop()
q := collection.NewQueue[string](); q.Enqueue("a"); v, _ = q.Dequeue()
pq := collection.NewPriorityQueue(func(a, b int) bool { return a < b })
pq.Push(3); pq.Push(1); min, _, _ := pq.Pop()
a, b := collection.NewSet(1, 2), collection.NewSet(2, 3)
r := collection.Intersection(a, b)
```
