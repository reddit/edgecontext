THRIFT_CMD=./thrift-docker.sh
THRIFT_GO_BASE_CMD=$(THRIFT_CMD) -out internal --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift,package_prefix=github.com/reddit/edgecontext/
GO=go
GO_TEST=$(GO) test -race ./...


.PHONY: thrift
thrift:
	cp ../../edgecontext.thrift .
	$(THRIFT_GO_BASE_CMD) edgecontext.thrift
	rm edgecontext.thrift
	find internal -depth -name "*-remote" -type d -exec rm -Rf {} \;


.PHONY: lint
lint:
	sh linters.sh
	go vet ./...


.PHONY: test
test:
	$(GO_TEST)


.PHONY: fmt
fmt:
	gofmt -s -w edgecontext/*.go


.PHONY: docs
docs:
	# nothing to do, just make sure we have the same make rules as python library.
