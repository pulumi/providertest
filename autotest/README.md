# AutoTest

AutoTest is a thin(ish) wrapper over the Automation API, making it easier to use within test scenarios.

The Automation API is just a thin wrapper around the calling the Pulumi CLI (`pulumi ...`). Some additional commands which are not yet supported by the Automation API are also added for convenience.

## Getting Started

The starting point to testing a program is to create a new AutoTest pointing at some existing program.

```go
func TestAProgram(t *testing.T) {
  source := NewAutoTest(t, filepath.Join("path", "to", "program"))
  //...
}
```

It's normally a good idea to run your program from a temporary directory to avoid cluttering your working directory with temporary or ephemeral files.

```go
source := NewAutoTest(t, ...)
program := source.CopyToTempDir()
```

The `source` variable is still pointing to the original path, but the `copy` is pointing to a new AutoTest in temporary directory which will automatically get removed at the end of the test.

Before we can preview or deploy a program we need to install dependencies and create a stack:

```go
var source *AutoTest
//...
program.Install() // Runs `pulumi install` to restore all dependencies
program.NewStack("my-stack") // Creates a new stack and returns it.
```

The created stack is returned but is also set as the current stack on the AutoTest object. All methods such as `source.Preview()` or `source.Up()` will use this current stack.

> [!NOTE]
> The new stack will be automatically destroyed and removed at the end of the test.

## Default Settings

`PULUMI_BACKEND_URL` is set to a temporary directory. This improves test performance and doesn't rely on the user already being authenticated to a specific backend account. This also isolates stacks so the same stack name can be re-used for several tests at once without risking conflicts and avoiding stack name randomisation which breaks importing & exporting between test runs. This can be overridden by setting `autoTest.Env().UseAmbientBackend()` or by setting `PULUMI_BACKEND_URL` yourself in the stack initialization options.

`PULUMI_CONFIG_PASSPHRASE` is set by default to "correct horse battery staple" (an arbitrary phrase) so that encrypted values are not tied to an external secret store that the user might not have access to. This can be overridden by setting `PULUMI_CONFIG_PASSPHRASE` in the stack initialization options.

## Configuring Providers

Pulumi discovers plugins the same as when running Pulumi commands directly.

In a test scenario, we often want to ensure a specific implementation of a provider is used during testing. The most reliable way is to configure use plugin attachment `PULUMI_DEBUG_PROVIDERS=NAME:PORT`. This prevents the Pulumi engine from searching for and starting the provider with the given name. Instead, it will connect to the already-running provider on the specified port. If the provider is not reachable on the given port, Pulumi will throw an error.

These can be specified via the `source.Env()` helpers:

```go
// Start a provider yourself
program.Env().AttachProvider("gcp", func(ctx context.Context) (int, error) {
  return port, nil // TODO: Actually start a provider.
})
// Start a server for testing from a pulumirpc.ResourceProviderServer implementation
program.Env().AttachProviderServer("gcp", func() (pulumirpc.ResourceProviderServer, error) {
  return makeProvider()
})
// Specify a local path where the binary lives to be started and attached.
program.Env().AttachProviderBinary("gcp", filepath.Join("..", "bin"))
// Use Pulumi to download a specific published version, then start and attach it.
program.Env().AttachDownloadedPlugin("gcp", "6.61.0")
```

## Common Operations

### Init

Copy the source to a temp directory, install dependencies and create a new stack:

```go
program := source.Init("my-stack")
```

### Update Source

Update source files for a subsequent step in the test:

```go
program.UpdateSource("folder_with_updates")
```

### Set Config

Set a variable in the stack's config:

```go
program.SetConfig("gcp:project", "pulumi-development")
```

### Pulumi Operations

Preview, Up, Refresh and Destroy:

```go
program.Preview()
program.Up()
program.Refresh()
program.Destroy()
```

> [!NOTE]
> Stacks created with `Init` or `NewStack` will be automatically destroyed and removed at the end of the test.

## Example

Here's a complete example as a test might look for the gcp provider with a local pre-built binary.

```go
func TestExample(t *testing.T) {
  // Copy test_dir to temp directory, install deps and create "my-stack"
  program := NewAutoTest(t, "test_dir").Init("my-stack")

  // Configure the test environment & project
  program.Env().AttachProviderBinary("gcp", "../bin")
  program.SetConfig("gcp:project", "pulumi-development")

  // Preview with assert
  preview := program.Preview()
  assert.Equal(t,
    map[apitype.OpType]int{apitype.OpCreate: 2},
    preview.ChangeSummary)

  // Up with assert
  deploy := program.Up()
  assert.Equal(t,
    map[string]int{"create": 2},
    *deploy.Summary.ResourceChanges)

  // Access logs for troubleshooting
  t.Log(deploy.StdOut)
}
```
