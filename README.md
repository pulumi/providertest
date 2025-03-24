# Pulumi Provider Testing Library

Library for testing Pulumi providers by running Pulumi programs.

> [!NOTE]
> The libraries in this repo are used internally by the Pulumi providers team, and are still evolving; you should expect incomplete documentation and breaking changes.

The library is composed of several modules. The most important of these is the [`pulumitest`](./pulumitest/) module. This is a library designed for testing any Pulumi program within a Go test. It extends the Go [Automation API](https://www.pulumi.com/automation/) with defaults appropriate for local testing such as using temporary directories for state.

Here's a short example of how to use pulumitest:

```go
import (
  "filepath"
  "github.com/pulumi/providertest/pulumitest"
)

func TestPulumiProgram(t *testing.T) {
  test := NewPulumiTest(t, filepath.Join("path", "to", "program"))
  test.Preview(t)
  test.Up(t)
  test.Refresh(t)
}
```

Refer to [the full documentation](./pulumitest/README.md) for a complete walkthrough of the features.

## Attaching In-Process Providers

If your provider implementation is available in the context of your test, the provider can be started in a background goroutine and used within the test using the `opttest.AttachProviderServer`. This avoids needing to build a provider binary before running the test, and allows stepping through from the test to the provider implementation when attaching a debugger.

For bridged providers using the standard repository layout, this can be configured as such:

```go
//go:embed cmd/pulumi-resource-example/schema.json
var schemaBytes []byte // Embed the generated schema (this might need to be re-generated before re-running tests)

func exampleResourceProviderServerFactory(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
  ctx := context.Background()
  version.Version = "1.0.0" // Set the global version to a non-empty string
  info := Provider() // Call the function defined in resource.go
  return tfbridge.NewProvider(ctx, nil, "example", version.Version, info.P, info, schemaBytes), nil
}
```

For native providers this function just returns your implementation of [the `pulumirpc.ResourceProviderServer` interface](https://pkg.go.dev/github.com/pulumi/pulumi/sdk/v3/proto/go#ResourceProviderServer).

## Upgrade Testing

We perform "upgrade testing" on providers to fail when a resource might be re-created when updating to a new version of the provider.

In the root `providertest` module there is a function called `PreviewProviderUpgrade(..)`. This shows the result of a preview when upgrading from a **baseline** version of a provider to a new version of the provider. On first run it records the *baseline* state after running the program with the *baseline* version of the provider and writes it into a `testdata` directory. On subsequent runs, it restores the state from the recorded *baseline* before performing a preview operation with the new version.

Here's an example of how to write an upgrade test:

```go
pt := pulumitest.NewPulumiTest(t, "path-to-a-pulumi-program-dir",
  // Use our local implementation for the new version
  opttest.AttachProviderServer("my-provider-name", exampleResourceProviderServerFactory))
// Perform a preview of upgrading from v0.0.1 of my-provider-name to our new version.
previewResult := providertest.PreviewProviderUpgrade(t, pt, "my-provider-name", "0.0.1")
// Assert the preview shows no changes
assertpreview.HasNoChanges(t, previewResult)
```

It's expected that the preview operation does not perform actual network calls, though it might still require credentials to be present for the provider's `Configure` method. Where the program under test calls invokes which might fail if the original test resources no longer exist, we can intercept the invokes and replay the original responses from the gRPC messages recorded at the same time as the recorded baseline state:

```go
// Turn a server factory into a *resource provider* server factory
resourceProviderFactory := providers.ResourceProviderFactory(exampleResourceProviderServerFactory)
// Calculate the path to the baseline version recording
upgradeCacheDir := providertest.GetUpgradeCacheDir("path-to-a-pulumi-program-dir", "0.0.1")
// Create a new factory which will intercept and replay invokes from the recorded grpc.json
factoryWithReplay := resourceProviderFactory.ReplayInvokes(filepath.Join(upgradeCacheDir, "grpc.json"), true)

pt := pulumitest.NewPulumiTest(t, "path-to-a-pulumi-program-dir",
  // Use the wrapped version that will intercept invokes
  opttest.AttachProviderServer("my-provider-name", factoryWithReplay))
```

## Other Modules

The `providers` module provides additional utilities for `pulumitest` when building providers:

- Attaching a running provider for a specific test.
- Starting a resource provider from within the same Go package so it can be attached and stepped through using a debugger within the test.
- Using a specific local provider binary.
- Downloading provider plugins at specific versions.
- Creating a mock of a resource provider.
- Intercepting calls to a provider via a proxy provider.
- Replaying previously captured invoke calls from a file.

The `grpclog` module contains types and functions for reading, querying and writing Pulumi's grpc log format (normally living in a `grpc.json` file).

The `replay` module has methods for exercising specific provider gRPC methods directly and from existing log files.
