GO = go

PKG_LIST=logger metrics

LIB_LIST=logger metrics

.PHONY: dep
dep:
	@echo "go mod download"
	@for pkg in ${PKG_LIST} ; do \
		cd $$pkg && $(GO) mod download && cd ..; \
	done


test:
	@for pkg in ${PKG_LIST} ; do \
		echo "test $$pkg" && cd $$pkg && $(GO) test ./... -cover && cd ..; \
	done
