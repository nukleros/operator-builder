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
	find . -name ${TEST_SCRIPT} | xargs dirname | xargs -I {} cp -r {} $(TEST_PATH)/.workloadConfig
	cd $(TEST_PATH); basename ${TEST_SCRIPT} | xargs find ${TEST_PATH} -name | xargs sh

test-clean:
	rm -rf $(TEST_PATH)/*
	rm -rf $(TEST_PATH)/.workloadConfig

DEBUG_PATH ?= test/application

debug-clean:
	rm -rf $(DEBUG_PATH)/*

debug-init: debug-clean
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- init \
		--workload-config .workloadConfig/workload.yaml \
   		--repo github.com/acme/acme-cnp-mgr \
        	--skip-go-version-check
debug-create:
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- create api \
		--workload-config .workloadConfig/workload.yaml \
		--controller \
		--resource

debug: debug-init debug-create