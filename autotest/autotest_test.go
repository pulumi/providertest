package autotest

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestDeploy(t *testing.T) {
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
	yamlTest.ImportStack(yamlStack)

	yamlPreview2 := yamlTest.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpSame: 2},
		yamlPreview2.ChangeSummary)
}

func TestConvert(t *testing.T) {
	sourceTest := NewAutoTest(t, filepath.Join("testdata", "yaml_program"))

	// Convert the original source to Python.
	pythonTest := sourceTest.Convert("python").AutoTest
	assert.NotEqual(t, pythonTest.Source(), sourceTest.Source())

	pythonTest.Install()
	pythonTest.NewStack("test")

	pythonPreview := pythonTest.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		pythonPreview.ChangeSummary)

	pythonUp := pythonTest.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*pythonUp.Summary.ResourceChanges)

	// Show the deploy output.
	t.Log(pythonUp.StdOut)
}

func TestBinaryAttach(t *testing.T) {
	program := NewAutoTest(t, filepath.Join("testdata", "yaml_gcp")).Init("")

	program.Env().AttachDownloadedPlugin("gcp", "6.61.0")
	program.SetConfig("gcp:project", "pulumi-development")

	preview := program.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		preview.ChangeSummary)

	deploy := program.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*deploy.Summary.ResourceChanges)

	program.UpdateSource(filepath.Join("testdata", "yaml_gcp_updated"))
	update := program.Up()
	assert.Equal(t,
		map[string]int{"same": 1, "update": 1},
		*update.Summary.ResourceChanges)
}
