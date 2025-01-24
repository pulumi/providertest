# Provider Testing

Incubating facilities for testing Pulumi providers

> [!NOTE]
> The libraries in this repo are used internally by the Pulumi providers team, but are still evolving; you should expect incomplete documentation and breaking changes. If you do choose to use this library we strongly reccomend starting with [providertest/pulumitest](https://github.com/pulumi/providertest/tree/main/pulumitest) which is our latest approach. Code in this repository that is not in the pulumitest subdirectory is very likely to be deprecated and removed in the near future. 

## Test Modes

### End To End (e2e)

Purpose: Prove provider behaviour for a resource is correct.

This test mode does not use the SDKs, but executes the YAML program directly. There are two types of e2e test: quick and full:

1. Quick will not provision real resources with the cloud provider.
2. Full will test the whole lifecycle of resources.

### SDK

Purpose: Ensure parity of each supported language's SDK behaviour.

SDK test are therefore split out per-language. Internally, this uses `pulumi convert` to automatically create the language-specific programs before executing them.

### VerifyUpgrade

Purpose: Verifies that upgrading the provider does not generate any unexpected replacements.

What these tests specifically try to verify is that the provider binary release candidate under test
will not generate any surprises for users attempting to upgrade to it from the baseline released
version.

There are currently several
[UpgradeTestMode](https://github.com/search?q=repo%3Apulumi%2Fprovidertest+type+UpgradeTestMode&type=code)
variations tests can run under, with different speed/accuracy trade-offs.


## Example Usage

```go
func TestSimple(t *testing.T) {
  test := NewProviderTest("test/simple", // Point to a directory containing a Pulumi YAML program
    WithProvider(StartLocalProvider), // Provider can be started and attached in-process
    WithUpdateStep(UpdateStepDir("../simple-2"))) // Multi-step tests are supported
  test.Run(t)
}
```

When calling `.Run()`, a suite of nested tests are run:

- `TestSimple/e2e`
- `TestSimple/sdk-csharp`
- `TestSimple/sdk-go`
- `TestSimple/sdk-python`
- `TestSimple/sdk-typescript`

If you want to run just one of these test modes directly locally, then you can temporarily replace `.Run(t)` with:

```go
test.RunE2e(t, true /*runFullTest*/)
test.RunSdk(t, "nodejs" /*language*/)
```

### Upgrade Tests

Set these extra options to enable upgrade tests:

```go
func TestSimple(t *testing.T) {
  test := NewProviderTest("test/simple",
    WithProviderName("gcp"),
    WithBaselineVersion("6.67.0"),
    WithResourceProviderServer(...))
  test.Run(t)
}
```

These nested tests are added:

- `TestSimple/upgrade-snapshot`
- `TestSimple/upgrade-preview-only`
- `TestSimple/upgrade-quick`
- `TestSimple/upgrade-full`

Note that `upgrade-snapshot` is a utility job rather than a test. `go test --provider-snapshot` runs
this job to exercise the baseline version of the provider and record its behavior under `testdata`.
The resulting recorded snapshot files are currently expected to be checked into the repo. They are
used to inform `upgrade-quick` and `upgrade-preview-only` tests. When updating the baseline version,
snapshots need to be recorded anew on the new version.

### Fixing failing tests
- If the tests fail by flagging unwanted resource updates or replacements that are actually
  acceptable, configure a custom
  [DiffValidation](https://github.com/pulumi/providertest/blob/5f23c3ec7cee882392ea356a54c0f74f56b0f7d5/upgrade.go#L241)
  setting with more relaxed asserts.

- If the tests flag legitimate upgrade issues, fixes are necessarily specific to the provider and resource being tested. 

- Remember to re-record the test snapshots when making changes to the example program or the
  baseline provider dependency.

## Controlling Test Mode

Which subtests are run, and in which mode (quick/full), are controlled by custom `go test` CLI flags. These can be set in makefiles or CI scripts as required.

To run all sub-tests:

```bash
go test -provider-e2e -provider-sdk-all ./...
```

By default, if no modes are explicitly set, only the fast end-to-end (e2e) sub-test is executed.

### Environment Variables

As a temporary control method, test mode can also be enabled via the environment variable `PULUMI_PROVIDER_TEST_MODE`. Multiple options can be specified separated by commas:

```env
PULUMI_PROVIDER_TEST_MODE=e2e,sdk-python
```

> [!NOTE]
> The environment variables should not be used in make files or CI configuration. Prefer using CLI flags for this.

### Reference

| Option         | CLI flag                   | Environment      | Description                                                           |
|----------------|----------------------------|------------------|-----------------------------------------------------------------------|
| Skip E2E       | `-provider-skip-e2e`       | `skip-e2e`       | Skip e2e provider tests                                               |
| Full E2E       | `-provider-e2e`            | `e2e`            | Enable full e2e provider tests, otherwise uses quick mode by default. |
| All SDK        | `-provider-sdk-all`        | `sdk-all`        | Enable all SDK tests                                                  |
| C# SDK         | `-provider-sdk-csharp`     | `sdk-csharp`     | Enable C# SDK tests                                                   |
| Python SDK     | `-provider-sdk-python`     | `sdk-python`     | Enable Python SDK tests                                               |
| Go SDK         | `-provider-sdk-go`         | `sdk-go`         | Enable Go SDK tests                                                   |
| Typescript SDK | `-provider-sdk-typescript` | `sdk-typescript` | Enable TypeScript SDK tests                                           |
| Snapshot       | `-provider-snapshot`       | `snapshot`       | Create snapshots for use with quick e2e tests                         |
