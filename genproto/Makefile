.PHONY: gen-go
gen-go:
	@echo "=== generate proto"
	protoc -I=./api \
			--go_out=./go
			--go_opt=paths=source_relative \
			./api/*/*/*.proto


.PHONY: proto-compile
proto-compile:
	protoc -I ./api \
			--go_out=./api \
			--go_opt=paths=source_relative \
			--custom_out=./go \
			--custom_opt=paths=source_relative \
			--plugin=protoc-gen-custom=$(GOPATH)/bin/protoc-gen-go-idl \
			./api/example/*/*.proto