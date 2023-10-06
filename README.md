# providertest

Incubating facilities for testing Pulumi providers

## Example Usage

```go
NewProviderTest("test/dir",
  WithProvider(StartLocalProvider),
  WithEditDir("../dir2", WithClean()))
```
