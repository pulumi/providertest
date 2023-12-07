package pulumitest

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestDeploy(t *testing.T) {
	test := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))

	// Ensure dependencies are installed.
	test.Install()
	// Create a new stack with auto-naming.
	test.NewStack("")
	// Test a preview.
	yamlPreview := test.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		yamlPreview.ChangeSummary)
	// Now do a real deploy.
	yamlUp := test.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*yamlUp.Summary.ResourceChanges)

	// Export the stack state.
	yamlStack := test.ExportStack()
	test.ImportStack(yamlStack)

	yamlPreview2 := test.Preview()
	assertpreview.HasNoChanges(t, yamlPreview2)
}

func TestConvert(t *testing.T) {
	// No need to copy the source, since we're not going to modify it.
	source := NewPulumiTestInPlace(t, filepath.Join("testdata", "yaml_program"))

	// Convert the original source to Python.
	converted := source.Convert("python").AutoTest
	assert.NotEqual(t, converted.Source(), source.Source())

	converted.Install()
	converted.NewStack("test")

	pythonPreview := converted.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		pythonPreview.ChangeSummary)

	pythonUp := converted.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*pythonUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, pythonUp)

	// Show the deploy output.
	t.Log(pythonUp.StdOut)
}

func TestBinaryAttach(t *testing.T) {
	test := NewPulumiTest(t,
		filepath.Join("testdata", "yaml_azure"),
		opttest.AttachDownloadedPlugin("azure-native", "2.21.0"))
	test.Init("")

	test.SetConfig("azure-native:location", "WestUS2")

	preview := test.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 3},
		preview.ChangeSummary)

	deploy := test.Up()
	assert.Equal(t,
		map[string]int{"create": 3},
		*deploy.Summary.ResourceChanges)

	test.UpdateSource(filepath.Join("testdata", "yaml_azure_updated"))
	update := test.Up()
	assert.Equal(t,
		map[string]int{"same": 2, "update": 1},
		*update.Summary.ResourceChanges)
}
