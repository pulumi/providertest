MAKEFLAGS    := --warn-undefined-variables

# Test runner. Defaults to plain `go test` for local runs; CI overrides this
# with gotestsum to emit GitHub Actions annotations. Everything after `--` is
# forwarded to `go test`, so the test flags below are unchanged either way.
GO_TEST_EXEC ?= go test

.PHONY: test
test:
	$(GO_TEST_EXEC) -v -coverprofile="coverage.txt" -coverpkg=./... ./...
