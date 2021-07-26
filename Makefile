# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build:
	go build -o bin/operator-builder cmd/operator-builder/main.go

test-install: build
	go test -cover -coverprofile=./bin/coverage.out ./...
	sudo cp bin/operator-builder /usr/local/bin/operator-builder

test-coverage-view: test-install
	go tool cover -html=./bin/coverage.out	

TEST_PATH ?= /tmp
TEST_SCRIPT ?= default.sh

test: test-install
	mkdir $(TEST_PATH)/.test
	cp test/$(TEST_SCRIPT) $(TEST_PATH)/.test/
	(cd $(TEST_PATH); ./.test/$(TEST_SCRIPT))

test-clean:
	rm -rf $(TEST_PATH)/*
	rm -rf $(TEST_PATH)/.test

