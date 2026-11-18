MAKEFLAGS    := --warn-undefined-variables

# Test runner. Defaults to plain `go test` for local runs; CI overrides this
# with gotestsum to emit GitHub Actions annotations. Everything after `--` is
# forwarded to `go test`, so the test flags below are unchanged either way.
GO_TEST_EXEC ?= go test

# Tests run in parallel and several of them run `dotnet build` through
# `pulumi`. With .NET SDK 11 those builds share one MSBuild server process,
# started by whichever build runs first. It is a child of that test's language
# host, so it is killed when that `pulumi` command exits, which cancels the
# builds of other tests still using it ("Build FAILED" with 0 errors). Give
# each build its own MSBuild process instead.
export DOTNET_CLI_USE_MSBUILD_SERVER := 0

.PHONY: test
test:
	$(GO_TEST_EXEC) -v -coverprofile="coverage.txt" -coverpkg=./... ./...
