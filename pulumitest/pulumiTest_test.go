package pulumitest

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	source := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to Python.
	converted := source.Convert("python").PulumiTest
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
	test.InstallStack("my-stack")

	test.SetConfig("azure-native:location", "WestUS2")

	preview := test.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 3},
		preview.ChangeSummary)
}

func TestBinaryPlugin(t *testing.T) {
	gcpBinary, err := providers.DownloadPluginBinary("gcp", "7.2.1")
	require.NoError(t, err)
	test := NewPulumiTest(t,
		filepath.Join("testdata", "yaml_gcp"),
		opttest.LocalProviderPath("gcp", gcpBinary))
	test.InstallStack("my-stack")

	test.SetConfig("gcp:project", "pulumi-development")

	preview := test.Preview()
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		preview.ChangeSummary)

	deploy1 := test.Up()
	assert.Equal(t,
		map[string]int{"create": 2},
		*deploy1.Summary.ResourceChanges)

	test.UpdateSource("testdata", "yaml_gcp_updated")
	deploy2 := test.Up()
	assert.Equal(t,
		map[string]int{"same": 1, "update": 1},
		*deploy2.Summary.ResourceChanges)
}
