.PHONY: protoc-compile
protoc-compile:

# 	protoc -I ./api \
  #			--openapiv2_out ./api \
  #			--openapiv2_opt logtostderr=true \
  #			--openapiv2_opt json_names_for_fields=false \
  #			--go_out ./api \
  #			--go_opt=paths=source_relative \
  #			--go-gin_out ./api \
  #			--go-gin_opt=paths=source_relative \
  #			api/product/app/v1/v1.proto

	protoc -I ./api \
			--go_out ./api \
			--go_opt=paths=source_relative \
			--custom_out=./api \
			--custom_opt=paths=source_relative \
			--plugin=protoc-gen-custom=$(GOPATH)/bin/protoc-gen-go-idl \
			./api/example/*/*.proto


.PHONY: dep
dep:
	go mod tidy -v


.PHONY: run
run:
	go run main.go

