# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Run all tests: `go test -v -coverprofile="coverage.txt" -coverpkg=./... ./...`
- Run a single test: `go test -v github.com/pulumi/providertest/pulumitest -run TestName`
- Run tests in a specific package: `go test -v github.com/pulumi/providertest/pulumitest`
- Run with race detection: `go test -race ./...`
- Check for lint issues: `golangci-lint run`
- Use Makefile shortcut: `make test` (runs full test suite with coverage)

## Architecture

This library is designed for testing Pulumi providers and programs. It consists of several key modules:

### Core Module: `pulumitest`
The primary testing module that extends Pulumi's Automation API with testing-specific defaults. Key components:
- `PulumiTest`: Main test harness that manages test lifecycle (copying programs, installing dependencies, creating stacks)
- `PT`: Testing interface that wraps Go's `testing.T` with additional test-specific methods
- Functional options pattern via `opttest.Option` for configuration (e.g., `AttachProvider`, `TestInPlace`, `SkipInstall`)
- Operations: `Up()`, `Preview()`, `Refresh()`, `Destroy()`, `Import()` - all accept the test context and options

### Provider Module: `providers`
Utilities for managing providers during tests:
- `ProviderFactory`: Function type that starts providers and returns listening ports
- `ResourceProviderServerFactory`: Creates `pulumirpc.ResourceProviderServer` instances for in-process testing
- Provider attachment modes: attach in-process servers, use local binaries, download specific versions
- `StartProviders()`: Manages multiple provider instances and their lifecycle via context cancellation

### Upgrade Testing
Core functionality in root module for testing provider version upgrades:
- `PreviewProviderUpgrade()`: Records baseline state with old provider version, then previews with new version
- Uses testdata cache directory pattern: `testdata/recorded/TestProviderUpgrade/{programName}/{baselineVersion}`
- Supports replaying invoke calls from recorded gRPC logs to avoid external dependencies during preview

### Additional Modules
- `grpclog`: Reading/writing Pulumi's gRPC log format (grpc.json files) - includes types for parsing and manipulating gRPC logs, with sanitization support for secrets
- `replay`: Exercising provider gRPC methods directly from log files:
  - `Replay()`: Executes a single gRPC request against a provider and asserts response matches expected pattern
  - `ReplaySequence()`: Replays multiple gRPC events in order from a JSON array
  - `ReplayFile()`: Replays all ResourceProvider events from a PULUMI_DEBUG_GRPC log file
  - Uses pattern matching (e.g., "*" wildcards) to handle non-deterministic responses
- `optproviderupgrade`: Options for `PreviewProviderUpgrade()` including cache directory templates and baseline options
- `optrun`: Options for the `Run()` method, including caching and option layering
- `optnewstack`: Options for stack creation, including auto-destroy configuration
- Assertion modules (`assertup`, `assertpreview`, `assertrefresh`): Functions for asserting operation results like `HasNoChanges()`, `HasNoDeletes()`

## Testing Patterns
- Tests use temporary directories by default (copied from source with `CopyToTempDir()`)
- Stack state stored locally in temporary directories with fixed passphrase for deterministic encryption
- Use `opttest.TestInPlace()` to run tests without copying (for performance or specific requirements)
- Helper functions call `t.Helper()` to ensure correct test failure line numbers
- Context cancellation via `t.Cleanup()` ensures proper resource cleanup

## Code Style
- Imports: Standard library first, third-party second, grouped by package source with blank line separators
- Functions: CamelCase for exported, camelCase for unexported; descriptive verb-prefixed names
- Variables: camelCase with descriptive names; single letters only for short-lived scopes
- Types: Custom types defined at package level with clear purpose; interfaces with method comments
- Testing: Use t.Parallel() for concurrent tests; use sub-tests with t.Run() for test organization
- Documentation: Add comments for exported functions, types, and variables

## Error Handling
- Return errors as last return value
- Check errors immediately after function calls
- Use fmt.Errorf for error wrapping with context
- Early returns on error conditions
- Helper functions for common error handling patterns

## Pull Requests
- Keep PRs focused on a single concern
- Add tests for new functionality
- Ensure tests pass before submission
- Follow existing code patterns when adding new functionality