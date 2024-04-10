package pulumitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/providertest/pulumitest/optrun"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// Run will run the `execute` function in an isolated temp directory and with additional test options, then import the resulting stack state into the original test.
// WithCache can be used to skip executing the run and return the cached stack state if available, or to cache the stack state after executing the run.
// Options will be inherited from the original test, but can be added to with `optrun.WithOpts` or reset with `opttest.Defaults()`.
func (pulumiTest *PulumiTest) Run(execute func(test *PulumiTest), opts ...optrun.Option) *PulumiTest {
	pulumiTest.PT().Helper()

	options := optrun.DefaultOptions()
	for _, o := range opts {
		o.Apply(options)
	}

	var stackExport *apitype.UntypedDeployment
	var err error
	if options.EnableCache {
		stackExport, err = tryReadStackExport(options.CachePath)
		if err != nil {
			pulumiTest.PT().Fatalf("failed to read stack export: %v", err)
		}
		if stackExport != nil {
			pulumiTest.PT().Logf("run cache found at %s", options.CachePath)
		} else {
			pulumiTest.PT().Logf("no run cache found at %s", options.CachePath)
		}
	}

	if stackExport == nil {
		isolatedTest := pulumiTest.CopyToTempDir(options.OptTest...)
		execute(isolatedTest)
		exportedStack := isolatedTest.ExportStack()
		if options.EnableCache {
			isolatedTest.PT().Logf("writing stack state to %s", options.CachePath)
			err = writeStackExport(options.CachePath, &exportedStack, false /* overwrite */)
			if err != nil {
				isolatedTest.PT().Fatalf("failed to write snapshot to %s: %v", options.CachePath, err)
			}
		}
		stackExport = &exportedStack
	}
	// Workaround: Previously recorded stack exports may contain randomised stack names which we need to fix before importing.
	// This can be removed once all old snapshots have been regenerated.
	stackName := pulumiTest.CurrentStack().Name()
	fixedStack, err := fixupStackName(stackExport, stackName)
	if err != nil {
		pulumiTest.PT().Fatalf("failed to fixup stack name: %v", err)
	}
	if fixedStack != stackExport {
		pulumiTest.PT().Logf("updating snapshot with fixed stack name: %s", stackName)
		err = writeStackExport(options.CachePath, fixedStack, true /* overwrite */)
		if err != nil {
			pulumiTest.PT().Fatalf("failed to write snapshot to %s: %v", options.CachePath, err)
		}
		stackExport = fixedStack
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

func fixupStackName(stateFile *apitype.UntypedDeployment, newStackName string) (*apitype.UntypedDeployment, error) {
	oldStackName, err := parseStackName(stateFile)
	if err != nil {
		return nil, err
	}
	if oldStackName == newStackName {
		return stateFile, nil
	}
	newDeployment := bytes.ReplaceAll([]byte(stateFile.Deployment), []byte(oldStackName), []byte(newStackName))
	newStateFile := *stateFile
	newStateFile.Deployment = json.RawMessage(newDeployment)
	return &newStateFile, nil
}

func parseStackName(state *apitype.UntypedDeployment) (string, error) {
	type Deployment struct {
		Resources []struct {
			URN  string `json:"urn"`
			Type string `json:"type"`
		} `json:"resources"`
	}
	var deployment Deployment
	err := json.Unmarshal([]byte(state.Deployment), &deployment)
	if err != nil {
		return "", err
	}
	var stackUrn string
	for _, r := range deployment.Resources {
		if r.Type == "pulumi:pulumi:Stack" {
			stackUrn = r.URN
			break
		}
	}
	return strings.Split(stackUrn, ":")[2], nil
}
