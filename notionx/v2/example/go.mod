module notionx-example

go 1.21

require github.com/tenz-io/gokit/notionx/v2 v2.0.0

require (
	github.com/tenz-io/notionapi v1.0.3 // indirect
	github.com/yuin/goldmark v1.7.8 // indirect
)

replace github.com/tenz-io/gokit/notionx/v2 => ./..
