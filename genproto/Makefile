.PHONY: proto-compile
proto-compile:
	@echo "=== generate proto"
	protoc -I=./api \
			--go_out=./go \
			--go_opt=paths=source_relative \
			./api/*/*/*.proto

