GO = go

.PHONY: test
test:
	$(GO) test ./... -cover -v

.PHONY: cover
cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	gofmt -w *.go

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -f coverage.out coverage.html

# Run every example program under example/.
.PHONY: run-example
run-example:
	@for d in example example-*; do \
		[ -d "$$d" ] || continue; \
		echo "==> $$d"; \
		(cd "$$d" && $(GO) run .) || exit 1; \
	done
