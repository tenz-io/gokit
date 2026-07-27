module cache-v3-example

go 1.24

require github.com/tenz-io/gokit/cache/v3 v3.0.0

require (
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)

replace github.com/tenz-io/gokit/cache/v3 => ./..
