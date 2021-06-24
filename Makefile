build:
	$(GOROOT)/bin/go build -o bin/operator-builder cmd/main.go

test-install: build
	$(GOROOT)/bin/go test -cover -coverprofile=./bin/coverage.out ./...
	sudo cp bin/operator-builder /usr/local/bin/operator-builder

test-coverage-view: test-install
	$(GOROOT)/bin/go tool cover -html=./bin/coverage.out	

TEST_PATH ?= /tmp
TEST_SCRIPT ?= default.sh

test: test-install
	mkdir $(TEST_PATH)/.test
	cp test/$(TEST_SCRIPT) $(TEST_PATH)/.test/
	(cd $(TEST_PATH); ./.test/$(TEST_SCRIPT))

test-clean:
	rm -rf $(TEST_PATH)/*
	rm -rf $(TEST_PATH)/.test

