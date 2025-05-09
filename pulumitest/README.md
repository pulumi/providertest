# Pulumi Testing Library

Pulumi Test is a thin(ish) wrapper over the Automation API, making it easier to use within test scenarios.

The Automation API is just a thin wrapper around the calling the Pulumi CLI (`pulumi ...`). Some additional commands which are not yet supported by the Automation API are also added for convenience.

## Getting Started

The starting point to testing a program is to create a new PulumiTest pointing at some existing test.

```go
func TestPulumiProgram(t *testing.T) {
  test := NewPulumiTest(t, filepath.Join("path", "to", "program"))
  //...
}
```

By default, your program is copied to a temporary directory before running to avoid modifying your source files or cluttering your working directory with temporary files. This will use an OS-specific temporary location by default but can be set to a custom directory using the `opttest.TempDir(dir)` option or the `PULUMITEST_TEMP_DIR` environment variable. Using an ignored directory within your repository can be useful for being able to locate any left-over folders retained from failed tests:

```go
test := NewPulumiTest(t, filepath.Join("path", "to", "program"), opttest.TempDir(".temp"))
```

If you don't want to copy your program to a temporary directory, use `opttest.TestInPlace()`. Also, you can also do a copy of a test manually by calling `CopyToTempDir()`:

```go
source := NewPulumiTest(t, opttest.TestInPlace())
copy := source.CopyToTempDir(t)
```

The `source` variable is still pointing to the original path, but the `copy` is pointing to a new PulumiTest in temporary directory which will automatically get removed at the end of the test.

The default behaviour is also to install dependencies and create a new stack called "test".

- The default stack name can be customised by using `opttest.StackName("my-stack")`.
- To prevent the automatic install of dependencies use `opttest.SkipInstall()`.
- To prevent automatically creating a stack use `opttest.SkipStackCreate()`

The following program is equivalent to the default test setup:

```go
test := NewPulumiTest(t, opttest.SkipInstall(), opttest.SkipStackCreate())
test.Install(t) // Runs `pulumi install` to restore all dependencies
test.NewStack(t, "test") // Creates a new stack and returns it.
```

The `Install` and `NewStack` steps can also be done together by calling `InstallStack()`:

```go
test.InstallStack("test")
```

The created stack is returned but is also set as the current stack on the PulumiTest object. All methods such as `source.Preview()` or `source.Up()` will use this current stack.

> [!NOTE]
> The new stack will be automatically destroyed and removed at the end of the test.

## Default Settings

`PULUMI_BACKEND_URL` is set to a temporary directory. This improves test performance and doesn't rely on the user already being authenticated to a specific backend account. This also isolates stacks so the same stack name can be re-used for several tests at once without risking conflicts and avoiding stack name randomisation which breaks importing & exporting between test runs. This can be overridden by the option `opttest.UseAmbientBackend()` or by setting `PULUMI_BACKEND_URL` yourself in the stack initialization options.

`PULUMI_CONFIG_PASSPHRASE` is set by default to "correct horse battery staple" (an arbitrary phrase) so that encrypted values are not tied to an external secret store that the user might not have access to. This can be overridden by setting `PULUMI_CONFIG_PASSPHRASE` in the stack initialization options.

When a test fails, the stack will attempt to be destroyed, though the temporary directories will remain in place. If you want to retain any resources which were created, you can set the env variable `PULUMITEST_SKIP_DESTROY_ON_FAILURE=true`.

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
test.Preview(t)
test.Up(t)
test.Refresh(t)
test.Destroy(t)
```

> [!NOTE]
> Stacks created with `InstallStack` or `NewStack` will be automatically destroyed and removed at the end of the test.

## Using Local SDKs

When running tests via SDKs that haven't yet been published, we need to configure the program under test to use our local build of the SDK instead of installing a version from their package registry.

### Node.js (Yarn Link)

For Node.js, we support using a [locally linked](https://classic.yarnpkg.com/lang/en/docs/cli/link/) version of the NPM package. Before running your test, you must run `yarn link` in the directory of the built Node.js SDK (normally in `sdk/nodejs/bin`).

Once the local link is configured, you can configure your test to use the linked package by using the `YarnLink` test option:

```go
NewPulumiTest(t, "test_dir", opttest.YarnLink("@pulumi/azure-native"))
```

### Go - Module Replacement

In Go, we support adding a replacement to the go.mod of the program under test. This is implemented by calling [`go mod edit -replace`](https://pkg.go.dev/cmd/go#hdr-Edit_go_mod_from_tools_or_scripts) with the user-specified replacement.

The replacement can be specified using the `GoModReplacement` test option:

```go
NewPulumiTest(t, "test_dir",
  opttest.GoModReplacement("github.com/pulumi/pulumi-my-provider/sdk/v3", "..", "sdk"))
```

## Additional Operations

### Update Source

Update source files for a subsequent step in the test:

```go
test.UpdateSource(t, "folder_with_updates")
```

### Set Config

Set a variable in the stack's config:

```go
test.SetConfig(t, "gcp:project", "pulumi-development")
```

## Environment Variables

The behavior of pulumitest can be adjusted through use of certain environment variables:

| Environment Variable | Purpose |
|----------------------|---------|
| `CI` | We inspect the `CI` environment flag to adjust certain defaults for an automated test environment |
| `PULUMITEST_RETAIN_FILES` | Set to `true` to always retain temporary files. |
| `PULUMITEST_RETAIN_FILES_ON_FAILURE` | Can be set explicitly to `true` or `false`. Defaults to `true` locally and `false` in CI environments. |
| `PULUMITEST_SKIP_DESTROY_ON_FAILURE` | Skips the automatic attempt to destroy a stack even after a test failure. This defaults to `false`. If set to true, the files will also be retained unless `PULUMITEST_RETAIN_FILES_ON_FAILURE` set to `false`. |
| `PULUMITEST_TEMP_DIR` | Changes the default temp directory from the OS-specific system location. |

## Asserts

The `assertup` and `assertpreview` modules contain a selection of functions for asserting on the results of the automation API:

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
  test := NewPulumiTest(t, "test_dir", opttest.AttachProviderBinary("gcp", "../bin"))
  test.InstallStack(t, "my-stack")

  // Configure the test environment & project
  test.SetConfig(t, "gcp:project", "pulumi-development")

  // Preview, Deploy, Refresh, 
  preview := test.Preview(t)
  t.Log(preview.StdOut)

  deploy := test.Up(t)
  t.Log(deploy.StdOut)
  assertpreview.HasNoChanges(t, test.Preview(t))

  // Export import
  test.ImportStack(t, test.ExportStack(t))
  assertpreview.HasNoChanges(t, test.Preview())

  test.UpdateSource(filepath.Join("testdata", "step2"))
  update := test.Up(t)
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
