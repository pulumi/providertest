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
copy := source.CopyToTempDir()
```

The `source` variable is still pointing to the original path, but the `copy` is pointing to a new AutoTest in temporary directory which will automatically get removed at the end of the test.

Before we can preview or deploy a program we need to install dependencies and create a stack:

```go
var source *AutoTest
//...
source.Install() // Runs `pulumi install` to restore all dependencies
source.NewStack("my-stack") // Creates a new stack and returns it.
```

The created stack is returned but is also set as the current stack on the AutoTest object. All methods such as `source.Preview()` or `source.Up()` will use this current stack.

## Default Settings

`PULUMI_BACKEND_URL` is set to a temporary directory. This improves test performance and doesn't rely on the user already being authenticated to a specific backend account. This also isolates stacks so the same stack name can be re-used for several tests at once without risking conflicts and avoiding stack name randomisation which breaks importing & exporting between test runs. This can be overridden by setting `autoTest.Env().UseAmbientBackend()` or by setting `PULUMI_BACKEND_URL` yourself in the stack initialization options.

`PULUMI_CONFIG_PASSPHRASE` is set to "correct horse battery staple" so that encrypted values are not tied to an external secret store that the user might not have access to. This can be overridden by setting `PULUMI_CONFIG_PASSPHRASE` in the stack initialization options.

## Configuring Providers

Pulumi discovers plugins the same as with normal execution. To ensure a specific implementation of a provider is used during testing we configure `PULUMI_DEBUG_PROVIDERS` with provider which are already running on a port. These can be specified via the `source.Env()` helper:

```go
source.Env().AttachDownloadedPlugin("gcp", "6.61.0")
source.Env().AttachProviderBinary("gcp", filepath.Join("..", "bin"))
source.Env().AttachProviderServer("gcp", func() (pulumirpc.ResourceProviderServer, error) {
  return makeProvider()
})
```

Note: If the provider is not reachable on the given port, Pulumi will throw an error.
