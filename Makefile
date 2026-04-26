GO = go

SUBMODULES := $(shell grep './' go.work | sed 's/^[ \t]*//' | grep -v '^use (' | grep -v '^)' | tr -d '\r' | awk '{ print $$1 }')


.PHONY: dep
dep:
	@echo "go mod tidy"
	@for module in $(SUBMODULES); do \
  		echo "pwd: $(shell pwd) && cd $$module && Tidying..."; \
		cd $$module && $(GO) mod tidy -v && cd - || exit 1; \
	done


test:
	@for module in $(SUBMODULES); do \
		echo "pwd: $(shell pwd) && cd $$module && Testing..."; \
		cd $$module && $(GO) test ./... -cover -v && cd - || exit 1; \
	done


gci:
	@echo "gci format"
	@for module in $(SUBMODULES); do \
  		echo "pwd: $(shell pwd) && cd $$module && Formating..."; \
		cd $$module && gci write -s standard -s default -s "prefix(github.com)" -s "prefix(github.com/tenz-io/gokit)" --skip-generated * && cd - || exit 1; \
	done


# Usage: make release VERSION=v2.0.5
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v2.0.5"; \
		echo "  make release VERSION=v2.0.5 DRY_RUN=1    # preview only"; \
		echo "  make release VERSION=v2.0.5 RELEASE=1     # also create GitHub Releases"; \
		exit 1; \
	fi
	@echo "=== Running tests ==="
	@$(MAKE) test
	@echo "=== Creating tags for version $(VERSION) ==="
	@if [ "$(DRY_RUN)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) --dry-run; \
	elif [ "$(RELEASE)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) --release; \
	else \
		./scripts/tag-all.sh $(VERSION) --push; \
	fi
	@echo "=== Done ==="
