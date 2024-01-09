package pulumitest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/optrun"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// Run will run the `execute` function in an isolated temp directory and with additional test options, then import the resulting stack state into the original test.
// WithCache can be used to skip executing the run and return the cached stack state if available, or to cache the stack state after executing the run.
// Options will be inherited from the original test, but can be added to with `optrun.WithOpts` or reset with `opttest.Defaults()`.
func (pulumiTest *PulumiTest) Run(execute func(test *PulumiTest), opts ...optrun.Option) *PulumiTest {
	pulumiTest.T().Helper()

	options := optrun.DefaultOptions()
	for _, o := range opts {
		o.Apply(options)
	}

	var stackExport *apitype.UntypedDeployment
	var err error
	if options.EnableCache {
		stackExport, err = tryReadStackExport(options.CachePath)
		if err != nil {
			pulumiTest.T().Fatalf("failed to read stack export: %v", err)
		}
		if stackExport != nil {
			pulumiTest.T().Logf("run cache found at %s", options.CachePath)
		} else {
			pulumiTest.T().Logf("no run cache found at %s", options.CachePath)
		}
	}

	if stackExport == nil {
		isolatedTest := pulumiTest.CopyToTempDir(options.OptTest...)
		execute(isolatedTest)
		exportedStack := isolatedTest.ExportStack()
		if options.EnableCache {
			isolatedTest.T().Logf("writing stack state to %s", options.CachePath)
			err = writeStackExport(options.CachePath, &exportedStack, false /* overwrite */)
			if err != nil {
				isolatedTest.T().Fatalf("failed to write snapshot to %s: %v", options.CachePath, err)
			}
		}
		stackExport = &exportedStack
	}
	pulumiTest.ImportStack(*stackExport)
	return pulumiTest
}

// writeStackExport writes the stack export to the given path creating any directories needed.
func writeStackExport(path string, snapshot *apitype.UntypedDeployment, overwrite bool) error {
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
	pathExists, err := exists(path)
	if err != nil {
		return err
	}
	if pathExists && !overwrite {
		return fmt.Errorf("stack export already exists at %s", path)
	}
	return os.WriteFile(path, stackBytes, 0644)
}

// tryReadStackExport reads a stack export from the given file path.
// If the file does not exist, returns nil, nil.
func tryReadStackExport(path string) (*apitype.UntypedDeployment, error) {
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
