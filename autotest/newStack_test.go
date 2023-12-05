package autotest

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestNewStackPreview(t *testing.T) {
	sourceTest := NewAutoTest(t, filepath.Join("testdata", "yaml_program"))

	// Test copying from the source directory to a temporary directory.
	yamlTest := sourceTest.CopyToTempDir()
	assert.NotEqual(t, yamlTest.Source(), sourceTest.Source())
	// Ensure dependencies are installed.
	yamlTest.Install()
	// Create a new stack with auto-naming.
	yamlTest.NewStack("")
	// Test a preview.
	yamlPreview := yamlTest.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		yamlPreview.ChangeSummary)
	// Now do a real deploy.
	yamlUp := yamlTest.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*yamlUp.Summary.ResourceChanges)

	// Export the stack state.
	yamlStack := yamlTest.ExportStack()

	// Convert the original source to Python.
	pythonTest := sourceTest.Convert("python").AutoTest
	assert.NotEqual(t, pythonTest.Source(), sourceTest.Source())
	// Also do a provider attach this time.
	pythonTest.Env().AttachDownloadedPlugin("gcp", "6.61.0")

	pythonTest.Install()
	// Use the same stack name as the YAML stack. It won't conflict because our state goes into an isolated temp directory.
	pythonTest.NewStack(yamlTest.CurrentStack().Name())
	// Re-import the YAML's deployed stack state.
	pythonTest.ImportStack(yamlStack)

	// Test that preview and up shows only sames.
	pythonPreview := pythonTest.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpSame: 2},
		pythonPreview.ChangeSummary)

	pythonUp := pythonTest.Up()
	assert.Equal(t, 1, len(*pythonUp.Summary.ResourceChanges))
	assert.Equal(t,
		map[string]int{"same": 2},
		*pythonUp.Summary.ResourceChanges)
	t.Log(pythonUp.StdOut)
}
