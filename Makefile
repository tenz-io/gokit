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


# Check version consistency across submodules before tagging.
.PHONY: version-check
version-check:
	@./scripts/version-check.sh


# Usage: make release VERSION=v3.0.1
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v3.0.1 [MODULES=mod1,mod2]"; \
		echo "  make release VERSION=v3.0.1 DRY_RUN=1   # preview tags"; \
		echo "  make release VERSION=v3.0.1 RELEASE=1  # also create GitHub Releases"; \
		exit 1; \
	fi
	@echo "=== Running tests ==="
	@$(MAKE) test
	@echo "=== Running version-check ==="
	@./scripts/version-check.sh
	@echo "=== Creating tags: version=$(VERSION) modules=$(MODULES) ==="
	@if [ "$(DRY_RUN)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) $(MODULES) --dry-run; \
	elif [ "$(RELEASE)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) $(MODULES) --release; \
	else \
		./scripts/tag-all.sh $(VERSION) $(MODULES) --push; \
	fi
	@echo "=== Done ==="
