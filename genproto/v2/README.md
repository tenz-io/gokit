# genproto

共享 protobuf 定义：Auth、RequestHeader、ResponseHeader。

```go
import "github.com/tenz-io/gokit/genproto/v2/go/custom/common"
```

## 快速开始

```go
import common "github.com/tenz-io/gokit/genproto/v2/go/custom/common"
header := &common.RequestHeader{Role: common.Role_USER, Userid: "123"}
```
