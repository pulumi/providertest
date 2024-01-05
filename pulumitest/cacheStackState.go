package pulumitest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// CacheStackState caches the stack state from executing the `run` func. Subsequent calls will load the stack state from the cache.
// If the cache does not exist, it will run the `run` func and cache the resulting stack state.
// The `run` function is always executed in an isolated temp directory so it won't affect the original test.
// The state exported from the `run` function (or cached from the previous run) is imported into the original test's stack.
func (pulumiTest *PulumiTest) CacheStackState(cachePath string, run func(test *PulumiTest), opts ...opttest.Option) *PulumiTest {
	pulumiTest.T().Helper()
	if cachePath == "" {
		pulumiTest.T().Fatal("cachePath is required")
	}

	stackExport, err := TryReadStackExport(cachePath)
	if err != nil {
		pulumiTest.T().Fatalf("failed to read stack export: %v", err)
	}
	if stackExport != nil {
		pulumiTest.T().Logf("load stack state from cache %s", cachePath)
	} else {
		pulumiTest.T().Logf("no stack state cache found at %s", cachePath)
		cacheRecording := pulumiTest.CopyToTempDir()
		run(cacheRecording)
		cacheRecording.T().Logf("writing stack state to %s", cachePath)
		exportedStack := cacheRecording.ExportStack()
		err = WriteStackExport(cachePath, &exportedStack)
		if err != nil {
			cacheRecording.T().Fatalf("failed to write snapshot to %s: %v", cachePath, err)
		}
		stackExport = &exportedStack
	}
	pulumiTest.ImportStack(*stackExport)
	return pulumiTest
}

// WriteStackExport writes the stack export to the given path creating any directories needed.
func WriteStackExport(path string, snapshot *apitype.UntypedDeployment) error {
	if snapshot == nil {
		return fmt.Errorf("stack export must not be nil")
	}
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	stackBytes, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, stackBytes, 0644)
}

// TryReadStackExport reads a stack export from the given file path.
// If the file does not exist, returns nil, nil.
func TryReadStackExport(path string) (*apitype.UntypedDeployment, error) {
	stackBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read stack export at %s: %v", path, err)
	}
	var stackExport apitype.UntypedDeployment
	err = json.Unmarshal(stackBytes, &stackExport)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stack export at %s: %v", path, err)
	}
	return &stackExport, nil
}
