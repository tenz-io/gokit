GO = go

PKG_LIST=logger monitor retriever app ginterceptor dbtracker httpcli

LIB_LIST=logger monitor retriever app ginterceptor dbtracker httpcli

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



gci:
	@echo "gci format"
	@for pkg in ${PKG_LIST} ; do \
		cd $$pkg && gci write -s standard -s default -s "prefix(github.com)" -s "prefix(github.com/tenz-io/gokit)" --skip-generated * && cd ..; \
	done

