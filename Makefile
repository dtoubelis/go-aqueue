PKG_VERSION?=0.0.0
COMMIT=`git rev-parse --short HEAD`
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

lint: ## Lint the files
	golint -set_exit_status $(PKG_LIST)

vet: ## Vet the file
	go vet $(PKG_LIST)

test:
	go test -short $(PKG_LIST)

race:
	go test -race -short $(PKG_LIST)

sec:
	gosec -exclude-dir .history -exclude-dir vendor ./...

coverage:
	./tools/coverage.sh

coverage_html:
	./tools/coverage.sh html

.PHONY: vet lint test coverage coverage_html race test sec
