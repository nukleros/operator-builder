# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
INIT_OPTS=init \
	--workload-config .workloadConfig/workload.yaml \
   	--repo github.com/acme/acme-cnp-mgr \
    --skip-go-version-check \
    --controller-image controller:latest \
	--enable-olm=true
CREATE_OPTS=create api \
	--workload-config .workloadConfig/workload.yaml \
	--enable-olm=true

define create_path
	if [ ! -d $(1)/.workloadConfig ]; then\
        mkdir -p $(1)/.workloadConfig;\
    fi
endef

export BASE_DIR := $(shell pwd)
export OPERATOR_BUILDER_PATH := $(BASE_DIR)/bin

.PHONY: build install test debug

build_command = go build -o bin/operator-builder cmd/operator-builder/main.go
build:
	$(call build_command)

install: build
	sudo cp bin/operator-builder /usr/local/bin/operator-builder

#
# traditional testing
#
GOLANGCI_LINT_VERSION ?= v1.57.1
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint:
	golangci-lint run

test:
	go test -cover -coverprofile=./bin/coverage.out ./...

test-coverage-view: test
	go tool cover -html=./bin/coverage.out

test-commit:
	test/scripts/commit-check-latest.sh

#
# debug testing with delve
#
TEST_WORKLOAD_PATH ?= test/cases/standalone

debug-clean:
	rm -rf $(TEST_WORKLOAD_PATH)/*

debug-init: debug-clean
	dlv debug ./cmd/operator-builder --wd $(TEST_WORKLOAD_PATH) -- $(INIT_OPTS)

debug-create:
	dlv debug ./cmd/operator-builder --wd $(TEST_WORKLOAD_PATH) -- $(CREATE_OPTS)

debug: debug-init debug-create

#
# simple functional code generation testing outside of codebase itself
#
FUNC_TEST_PATH ?= /tmp/operator-builder-func-test

func-test-clean:
	if [ -d $(FUNC_TEST_PATH) ]; then rm -rf $(FUNC_TEST_PATH)/*; fi

INIT_BUILD ?= true
func-test-init: func-test-clean
	if [ "$(INIT_BUILD)" == "true" ]; then $(call build_command); fi
	$(call create_path,$(FUNC_TEST_PATH))
	cp -r $(BASE_DIR)/$(TEST_WORKLOAD_PATH)/.workloadConfig/* $(FUNC_TEST_PATH)/.workloadConfig ;
	cd $(FUNC_TEST_PATH) && $(OPERATOR_BUILDER_PATH)/operator-builder $(INIT_OPTS)

func-test-create:
	cd $(FUNC_TEST_PATH) && $(OPERATOR_BUILDER_PATH)/operator-builder $(CREATE_OPTS)

func-test: func-test-init func-test-create