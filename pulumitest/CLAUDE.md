# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Run all tests in pulumitest module: `go test -v ./...`
- Run a single test: `go test -v -run TestName`
- Run with coverage: `go test -v -coverprofile="coverage.txt" -coverpkg=./... ./...`
- Run with race detection: `go test -race ./...`
- Check for lint issues: `golangci-lint run`
- From root: `go test -v github.com/pulumi/providertest/pulumitest`

## Architecture

The `pulumitest` module is the core testing library for Pulumi programs and providers. It wraps Pulumi's Automation API with testing-specific defaults and conveniences.

### Core Components

**PulumiTest** (`pulumiTest.go`)
- Main test harness that manages the full test lifecycle
- Handles program copying, dependency installation, stack creation/destruction
- Stores context, working directory, options, and current stack reference
- Created via `NewPulumiTest(t, source, ...opts)` with functional options pattern

**PT Interface** (`testingT.go`)
- Thin wrapper around Go's `testing.T` that provides test-specific methods
- All operations accept PT as first parameter for proper test reporting
- Helper methods call `t.Helper()` to ensure correct failure line numbers in test output

**Options System** (`opttest/opttest.go`)
- Functional options pattern via `opttest.Option` interface
- Key options: `AttachProvider`, `AttachProviderServer`, `AttachProviderBinary`, `TestInPlace`, `SkipInstall`, `SkipStackCreate`, `YarnLink`, `GoModReplacement`, `DotNetReference`, `LocalProviderPath`
- Options are deeply copied to allow independent modification when using `CopyToTempDir()`
- Default passphrase: "correct horse battery staple" for deterministic encryption

### Operations

**Stack Lifecycle** (`newStack.go`, `installStack.go`, `destroy.go`)
- `Install(t)`: Runs `pulumi install` to restore dependencies
- `NewStack(t, name, ...opts)`: Creates new stack with local backend by default, sets as current stack
- `InstallStack(t, name)`: Convenience combining Install + NewStack
- `Destroy(t)`: Destroys resources and removes stack (automatic via `t.Cleanup()`)
- Auto-destroy behavior configurable via `optnewstack.DisableAutoDestroy()` or `EnableAutoDestroy()`

**Pulumi Operations** (`up.go`, `preview.go`, `refresh.go`, `destroy.go`, `import.go`)
- `Up(t, ...opts)`: Runs `pulumi up`, returns `auto.UpResult`
- `Preview(t, ...opts)`: Runs `pulumi preview`, returns `auto.PreviewResult`
- `Refresh(t, ...opts)`: Runs `pulumi refresh`, returns `auto.RefreshResult`
- `Destroy(t)`: Destroys current stack
- `Import(t, resourceType, name, id, ...opts)`: Imports existing resource into state
- All operations accept `optrun.Option` for runtime configuration

**Test Utilities**
- `CopyToTempDir(t, ...opts)`: Creates independent copy in temp directory for isolated testing
- `UpdateSource(t, newSourcePath)`: Replaces program files with new version for multi-step tests
- `SetConfig(t, key, value)`: Sets stack configuration values
- `ExportStack(t)`: Exports stack state as deployment JSON
- `ImportStack(t, deployment)`: Imports stack state from deployment JSON
- `GrpcLog(t)`: Retrieves gRPC log for provider calls made during test
- `Run(t, fn, ...opts)`: Execute function with optional state caching and option layering

### Provider Attachment

The library supports multiple ways to configure providers for testing:

**Attach In-Process Server** (`opttest.AttachProviderServer`)
- Start `pulumirpc.ResourceProviderServer` implementation in-process via goroutine
- Allows debugging from test into provider code
- Factory function receives `PulumiTest` context

**Attach Local Binary** (`opttest.AttachProviderBinary`)
- Start pre-built provider binary and attach via `PULUMI_DEBUG_PROVIDERS`
- Path can be directory (assumes `pulumi-resource-<name>`) or full binary path

**Attach Downloaded Plugin** (`opttest.AttachDownloadedPlugin`)
- Downloads specific plugin version via `pulumi plugin install`
- Starts and attaches the downloaded provider

**Local Provider Path** (`opttest.LocalProviderPath`)
- Sets `plugins.providers` in Pulumi.yaml for providers that don't support attachment
- Provider is started by Pulumi engine, not attached

### Testing Patterns

**Default Behavior**
- Programs copied to temp directory (OS-specific or `PULUMITEST_TEMP_DIR`)
- Dependencies installed automatically unless `SkipInstall()` used
- Stack named "test" created automatically unless `SkipStackCreate()` used
- Local backend in temp directory unless `UseAmbientBackend()` used
- Stacks automatically destroyed on test completion
- Temp directories retained on failure for debugging (configurable via env vars)

**gRPC Logging** (`grpcLog.go`, `grpcLog_test.go`)
- Enabled by default, written to `grpc.json` in working directory
- Disable with `opttest.DisableGrpcLog()`
- Access via `test.GrpcLog(t)` which returns parsed log entries
- Supports sanitization of secrets before writing to disk

**Multi-Step Tests**
- Use `UpdateSource(t, path)` to replace program files between operations
- Useful for testing update behavior, replacements, etc.
- Example pattern: `Up()` → `UpdateSource()` → `Up()` → assert changes

**SDK Configuration**
- Node.js: Use `YarnLink("@pulumi/package")` after running `yarn link` in SDK directory
- Go: Use `GoModReplacement("module", "path", "to", "replacement")` to add go.mod replacements
- .NET/C#:
  - Use `DotNetReference("package", "path", "to", "project")` to add project references to .csproj files
    - Path can point to a .csproj file or a directory containing one
    - Absolute paths are resolved automatically
    - ProjectReference elements are added to the .csproj file on stack creation
  - Use `DotNetBuildConfiguration("Release")` to set build configuration (Debug/Release)
    - Sets `DOTNET_BUILD_CONFIGURATION` environment variable
    - Default is Debug if not specified
  - Use `DotNetTargetFramework("net8.0")` to override target framework
    - Modifies `<TargetFramework>` element in .csproj
    - Useful for testing across different .NET versions

**Examples**
```go
// Basic .NET test
test := NewPulumiTest(t, "path/to/csharp/project")
up := test.Up(t)

// Test with local SDK reference
test := NewPulumiTest(t,
    "path/to/csharp/project",
    opttest.DotNetReference("Pulumi.Aws", "../pulumi-aws/sdk/dotnet"),
)

// Test with specific framework and configuration
test := NewPulumiTest(t,
    "path/to/csharp/project",
    opttest.DotNetTargetFramework("net7.0"),
    opttest.DotNetBuildConfiguration("Release"),
)
```

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `CI` | Adjusts defaults for CI environment (affects file retention) |
| `PULUMITEST_RETAIN_FILES` | Set to `true` to always retain temp directories |
| `PULUMITEST_RETAIN_FILES_ON_FAILURE` | Retain temp files on test failure (default: `true` locally, `false` in CI) |
| `PULUMITEST_SKIP_DESTROY_ON_FAILURE` | Skip automatic destroy on test failure (default: `false`) |
| `PULUMITEST_TEMP_DIR` | Custom temp directory instead of OS default |
| `PULUMI_CONFIG_PASSPHRASE` | Override default passphrase (defaults to "correct horse battery staple") |
| `PULUMI_BACKEND_URL` | Override default local backend |

### Subdirectories

**opttest/** - Options for PulumiTest construction and stack creation
**optrun/** - Options for Up/Preview/Refresh/Destroy operations
**optnewstack/** - Options for NewStack (auto-destroy configuration)
**assertup/** - Assertions for Up results (`HasNoDeletes`, `HasNoChanges`, etc.)
**assertpreview/** - Assertions for Preview results
**assertrefresh/** - Assertions for Refresh results
**changesummary/** - Types for analyzing resource change summaries
**sanitize/** - Utilities for sanitizing sensitive data in logs and snapshots

### File Organization

- Core types: `pulumiTest.go`, `testingT.go`
- Operations: `up.go`, `preview.go`, `refresh.go`, `destroy.go`, `import.go`
- Stack management: `newStack.go`, `installStack.go`, `install.go`
- Utilities: `copy.go`, `updateSource.go`, `setConfig.go`, `exportStack.go`, `importStack.go`
- gRPC logging: `grpcLog.go`, `grpcLog_test.go`
- File system: `fs_unix.go`, `fs_windows.go`, `tempdir.go`
- Project file handling: `pulumiYAML.go`, `csproj.go`
- Command execution: `execCmd.go`, `run.go`
- Cleanup: `cleanup.go`

## Code Patterns

**Helper Functions**
- Always call `t.Helper()` in functions that accept PT to ensure correct test failure line numbers
- Return errors when operation can fail; panic only for programmer errors
- Accept variadic options as last parameters

**Testing**
- Use `t.Parallel()` for tests that can run concurrently
- Use `t.Run()` for sub-tests to organize test cases
- Cleanup via `t.Cleanup()` ensures resources freed even on test failure

**Context Management**
- Test context created from `t.Deadline()` if available
- Context cancelled via `t.Cleanup()` for automatic resource cleanup
- Provider factories receive context to handle graceful shutdown

**Options Pattern**
- All options implement `Option` interface with `Apply(*Options)` method
- Options are composable and order-independent (where possible)
- Use `Defaults()` to reset options to initial state
- Options deeply copied for independent test instances

## Troubleshooting

### .NET/C# Issues

**Target Framework Not Found**
- Error: `Framework 'Microsoft.NETCore.App', version 'X.X.X' not found`
- Solution: Install the required .NET SDK version or use `DotNetTargetFramework()` to specify an installed version
- Check installed versions: `dotnet --list-sdks`

**Build Configuration Issues**
- If builds are slow or producing unexpected results, explicitly set: `DotNetBuildConfiguration("Release")`
- Debug builds include more information but are larger and slower

**Project Reference Resolution**
- Ensure referenced projects use compatible target frameworks
- Use absolute paths or paths relative to the test working directory
- The test framework automatically resolves relative paths to absolute paths

**NuGet Package Version Mismatch**
- Pulumi .NET SDK versioning differs from CLI versioning
- Latest stable SDK: 3.90.0 (as of writing)
- Check NuGet.org for current versions

**Common Test Failures**
- `pulumi install` fails: Check .csproj package versions are available on NuGet
- Build fails with missing types: Verify all project references are correctly added
- Stack creation hangs: Check for `PULUMI_AUTOMATION_API_SKIP_VERSION_CHECK=true` in CI environments
