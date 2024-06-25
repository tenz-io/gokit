.PHONY: protoc-compile
protoc-compile:
	# go install
	protoc -I ./api \
			--go_out ./api \
			--go_opt=paths=source_relative \
			--go-validator_out ./api \
			--go-validator_opt=paths=source_relative \
			api/example/v1/example.proto

.PHONY: dep
dep:
	go mod tidy -v


.PHONY: run
run:
	go run main.go

