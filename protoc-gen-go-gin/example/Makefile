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


.PHONY: send-request
send-request:
	curl --verbose  localhost:8888/v1/author/123/articles/