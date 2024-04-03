GO = go

SUBMODULES := $(shell grep './' go.work | sed 's/^[ \t]*//' | grep -v '^use (' | grep -v '^)' | tr -d '\r' | awk '{ print $$1 }')


.PHONY: dep
dep:
	@echo "go mod tidy"
	@for module in $(SUBMODULES); do \
		cd $$module && $(GO) mod tidy -v && cd - || exit 1; \
	done


test:
	@for module in $(SUBMODULES); do \
		echo "Testing $$module..."; \
		cd $$module && $(GO) test ./... -cover -v && cd - || exit 1; \
	done



gci:
	@echo "gci format"
	@for module in $(SUBMODULES); do \
		cd $$module && gci write -s standard -s default -s "prefix(github.com)" -s "prefix(github.com/tenz-io/gokit)" --skip-generated * && cd - || exit 1; \
	done

