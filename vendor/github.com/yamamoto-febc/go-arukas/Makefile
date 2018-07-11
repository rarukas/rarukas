VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
CURRENT_VERSION = $(gobump show -r )
GO_FILES?=$(shell find . -name '*.go')

.PHONY: default
default: test

.PHONY: tools
tools:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/motemen/gobump
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: testacc
testacc: 
	TEST_ACC=1 go test ./... $(TESTARGS) -v -timeout=30m -parallel=4 ;

.PHONY: test
test: 
	TEST_ACC=  go test ./... $(TESTARGS) -v -timeout=30m -parallel=4 ;

.PHONY: lint
lint: fmt
	gometalinter --vendor --skip=vendor/ --cyclo-over=15 --disable=gas --deadline=2m ./...
	@echo

.PHONY: fmt
fmt:
	gofmt -s -l -w $(GOFMT_FILES)
