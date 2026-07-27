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


# One-shot: tag + GitHub Release for all modules via gh API.
# Usage: make release VERSION=v3.0.1
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v3.0.1 [MODULES=m1,m2] [REPO=owner/repo] [NOTES_FILE=path] [DRY_RUN=1] [PUSH=1]"; \
		echo "  make release VERSION=v3.0.1            # one-shot: tag + release all modules"; \
		echo "  make release VERSION=v3.0.1 DRY_RUN=1 # preview"; \
		echo "  make release VERSION=v3.0.1 PUSH=1    # tag only, no releases"; \
		echo "  make release VERSION=v3.0.1 REPO=tenz-io/gokit"; \
		echo "  Prereq: gh auth login"; \
		exit 1; \
	fi
	@echo "=== Running tests ==="
	@$(MAKE) test
	@echo "=== Running version-check ==="
	@./scripts/version-check.sh
	@echo "=== Creating tags + releases: version=$(VERSION) repo=$(REPO) modules=$(MODULES) ==="
	@if [ -n "$(NOTES_FILE)" ]; then NOTES_FLAG="--notes-from-file $(NOTES_FILE)"; else NOTES_FLAG=""; fi; \
	if [ "$(DRY_RUN)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) $(MODULES) $$NOTES_FLAG $${REPO:+--repo $(REPO)} --release --dry-run; \
	elif [ "$(PUSH)" = "1" ]; then \
		./scripts/tag-all.sh $(VERSION) $(MODULES) $$NOTES_FLAG $${REPO:+--repo $(REPO)} --push; \
	else \
		./scripts/tag-all.sh $(VERSION) $(MODULES) $$NOTES_FLAG $${REPO:+--repo $(REPO)} --release; \
	fi
	@echo "=== Done ==="
