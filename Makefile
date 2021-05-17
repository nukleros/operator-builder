build:
	go build -o bin/kbl cmd/main.go

test-install: build
	cp bin/kbl /usr/local/bin/kbl

TEST_PATH ?= /tmp
TEST_SCRIPT ?= default.sh
test: test-install
	mkdir $(TEST_PATH)/.test
	cp test/$(TEST_SCRIPT) $(TEST_PATH)/.test/
	(cd $(TEST_PATH); ./.test/$(TEST_SCRIPT))

test-clean:
	rm -rf $(TEST_PATH)/*
	rm -rf $(TEST_PATH)/.test

