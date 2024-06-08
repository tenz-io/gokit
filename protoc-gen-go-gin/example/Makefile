.PHONY: protoc-compile
protoc-compile:
	#go install
	protoc -I ./api \
	--openapiv2_out ./api --openapiv2_opt logtostderr=true \
	--openapiv2_opt json_names_for_fields=false \
	--go_out ./api --go_opt=paths=source_relative \
	--go-gin_out ./api --go-gin_opt=paths=source_relative \
	api/product/app/v1/v1.proto
	protoc-go-inject-tag -input=./api/product/app/v1/v1.pb.go


.PHONY: dep
dep:
	go mod tidy -v





.PHONY: get-articles
get-articles:
	curl -X GET "localhost:8888/v1/author/123/articles?title=hello&page=1&page_size=10"


.PHONY: create-article
send-request:
	curl -X POST localhost:8888/v1/author/123/articles \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer abc123" \
	-d '{"title":"Hello World","content":"This is a test article"}'