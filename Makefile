PKG_VERSION?=0.0.0
COMMIT=`git rev-parse --short HEAD`
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

lint: ## Lint the files
	golint -set_exit_status $(PKG_LIST)

vet: ## Vet the file
	go vet $(PKG_LIST)

sec:
	gosec -exclude-dir .history -exclude-dir vendor ./...

test:
	go test -v -benchmem -bench=. -short $(PKG_LIST)

race:
	go test -v -race -short $(PKG_LIST)

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: lint vet sec test race coverage
