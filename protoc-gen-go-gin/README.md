# protoc-gen-go-gin

> Modify from [varluffy/protoc-gen-go-gin](https://github.com/varluffy/protoc-gen-go-gin)

generate gin server code from proto file.

## Development

### Requirements

- [go 1.21](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)
- [protoc-gen-openapiv2](https://github.com/grpc-ecosystem/grpc-gateway/releases/tag/v2.19.1)
- [protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag)

Use embed feature in go 1.16, so go 1.16 is required.

```bash
make install

go install github.com/favadi/protoc-go-inject-tag@v1.4.0
```

## Usage

eg: [example](./example)

### proto pacification

默认情况下 rpc method 命名为 方法+资源，使用驼峰方式命名，生成代码时会进行映射

方法映射方式如下所示:

- `"GET", "FIND", "QUERY", "LIST", "SEARCH"`  --> GET
- `"POST", "CREATE"`  --> POST
- `"PUT", "UPDATE"`  --> PUT
- `"DELETE"`  --> DELETE

```protobuf
service BlogService {
  rpc CreateArticle(Article) returns (Article) {}
  // 生成 http 路由为 post: /article
}
```

除此之外还可以使用 google.api.http option 指定路由，可以通过添加 additional_bindings 使一个 rpc 方法对应多个路由

```protobuf
// blog service is a blog demo
service BlogService {
  rpc GetArticles(GetArticlesReq) returns (GetArticlesResp) {
    // 
    // 可以通过添加 additional_bindings 使一个 rpc 方法对应多个路由
    option (google.api.http) = {
      get: "/v1/articles"
      additional_bindings {
        get: "/v1/author/{author_id}/articles"
      }
    };
  }
}
```

### Generate code

```bash
cd example && make proto-comiple
```
