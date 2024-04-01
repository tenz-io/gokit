GO = go

PKG_LIST=logger monitor retriever app ginterceptor

LIB_LIST=logger monitor retriever app ginterceptor

.PHONY: dep
dep:
	@echo "go mod tidy"
	@for pkg in ${PKG_LIST} ; do \
		cd $$pkg && $(GO) mod tidy -v && cd ..; \
	done


test:
	@for pkg in ${PKG_LIST} ; do \
		echo "test $$pkg" && cd $$pkg && $(GO) test ./... -cover && cd ..; \
	done
