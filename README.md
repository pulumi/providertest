# providertest

Incubating facilities for testing Pulumi providers

## Example Usage

```go
NewProviderTest("test/dir",
  WithProvider(StartLocalProvider),
  WithEditDir("../dir2", WithClean()))
```

## Controlling Test Mode

| Option | CLI flag | Environment | Description |
|---|---|---|---|
| Skip E2E | `-provider-skip-e2e` | `skip-e2e` | Skip e2e provider tests |
| Full E2E | `-provider-e2e` | `e2e` | Enable full e2e provider tests, otherwise uses quick mode by default. |
| C# SDK | `-provider-sdk-csharp` | `sdk-csharp` | Enable C# SDK tests |
| Python SDK | `-provider-sdk-python` | `sdk-python` | Enable Python SDK tests |
| Go SDK | `-provider-sdk-go` | `sdk-go` | Enable Go SDK tests |
| Typescript SDK | `-provider-sdk-typescript` | `sdk-typescript` | Enable TypeScript SDK tests |
| Snapshot | `-provider-snapshot` | `snapshot` | Create snapshots for use with quick e2e tests |

The flags are set when executing the tests e.g.

```bash
go test -provider-e2e ./...
```

The environment options are set within `PULUMI_PROVIDER_TEST_MODE`. Multiple options can be specified - comma separated. E.g.

```env
PULUMI_PROVIDER_TEST_MODE=e2e,sdk-python
```
