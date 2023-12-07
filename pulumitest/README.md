# Pulumi Testing Library

Pulumi Test is a thin(ish) wrapper over the Automation API, making it easier to use within test scenarios.

The Automation API is just a thin wrapper around the calling the Pulumi CLI (`pulumi ...`). Some additional commands which are not yet supported by the Automation API are also added for convenience.

## Getting Started

The starting point to testing a program is to create a new AutoTest pointing at some existing test.

```go
func TestPulumiProgram(t *testing.T) {
  test := NewPulumiTest(t, filepath.Join("path", "to", "program"))
  //...
}
```

By default run your program is copied to a temporary directory before running to avoid cluttering your working directory with temporary or ephemeral files. To disable this behaviour, use `NewPulumiTestInPlace`. You can also do a copy of a test manually by calling `CopyToTempDir()`:

```go
source := NewAutoTestInPlace(t, ...)
copy := source.CopyToTempDir()
```

The `source` variable is still pointing to the original path, but the `copy` is pointing to a new AutoTest in temporary directory which will automatically get removed at the end of the test.

Before we can preview or deploy a program we need to install dependencies and create a stack:

```go
//...
test.Install() // Runs `pulumi install` to restore all dependencies
test.NewStack("my-stack") // Creates a new stack and returns it.
```

These two steps can also be done together by calling `InstallStack()`:

```go
test.InstallStack("my-stack")
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

These can be specified via the `Attach*` options when constructing the test:

```go
// Start a provider yourself
NewPulumiTest(t, "path", opttest.AttachProvider("gcp", func(ctx context.Context) (int, error) {
  return port, nil // TODO: Actually start a provider.
})
// Start a server for testing from a pulumirpc.ResourceProviderServer implementation
NewPulumiTest(t, "path", opttest.AttachProviderServer("gcp", func() (pulumirpc.ResourceProviderServer, error) {
  return makeProvider()
})
// Specify a local path where the binary lives to be started and attached.
NewPulumiTest(t, "path", opttest.AttachProviderBinary("gcp", filepath.Join("..", "bin"))
// Use Pulumi to download a specific published version, then start and attach it.
NewPulumiTest(t, "path", opttest.AttachDownloadedPlugin("gcp", "6.61.0")
```

For providers which don't support attaching, we can configure the path to the binary of a specific provider in the `plugins.providers` property in the project settings (Pulumi.yaml) by using the `LocalProviderPath()` option:

```go
NewPulumiTest(t, "path", opttest.LocalProviderPath("gcp", filepath.Join("..", "bin"))
```

## Pulumi Operations

Preview, Up, Refresh and Destroy can be run directly from the test context:

```go
test.Preview()
test.Up()
test.Refresh()
test.Destroy()
```

> [!NOTE]
> Stacks created with `InstallStack` or `NewStack` will be automatically destroyed and removed at the end of the test.

## Additional Operations

### Update Source

Update source files for a subsequent step in the test:

```go
test.UpdateSource("folder_with_updates")
```

### Set Config

Set a variable in the stack's config:

```go
test.SetConfig("gcp:project", "pulumi-development")
```

## Asserts

In parallel to the `autotest` module, the `autoassert` module contains a selection of functions for asserting on the results of the automation API:

```go
assertup.HasNoDeletes(t, upResult)
assertup.HasNoChanges(t, upResult)
assertpreview.HasNoChanges(t, previewResult)
assertpreview.HasNoDeletes(t, previewResult)
```

## Example

Here's a complete example as a test might look for the gcp provider with a local pre-built binary.

```go
func TestExample(t *testing.T) {
  // Copy test_dir to temp directory, install deps and create "my-stack"
  test := NewAutoTest(t, "test_dir", opttest.AttachProviderBinary("gcp", "../bin"))
  test.InstallStack("my-stack")

  // Configure the test environment & project
  test.SetConfig("gcp:project", "pulumi-development")

  // Preview, Deploy, Refresh, 
  preview := test.Preview()
  t.Log(preview.StdOut)

  deploy := test.Up()
  t.Log(deploy.StdOut)
  assertpreview.HasNoChanges(t, test.Preview())

  // Export import
  test.ImportStack(test.ExportStack())
  assertpreview.HasNoChanges(t, test.Preview())

  test.UpdateSource(filepath.Join("testdata", "step2"))
  update := test.Up()
  t.Log(update.StdOut)
}
```

Comparative ProgramTest example:

```go
func TestExample(t *testing.T) {
  test := integration.ProgramTestOptions{
    Dir: testDir(t, "test_dir"),
    Dependencies: []string{filepath.Join("..", "sdk", "python", "bin")},
    ExpectRefreshChanges: true,
    Config: map[string]string{
      "gcp:project": "pulumi-development",
    },
    LocalProviders: []integration.LocalDependency{
      {
        Package: "gcp",
        Path:    "../bin",
      },
    },
    EditDirs: []integration.EditDir{
      {
        Dir:      testDir(t, "test_dir", "step2"),
        Additive: true,
      },
    },
  }

  integration.ProgramTest(t, &test)
}
```
