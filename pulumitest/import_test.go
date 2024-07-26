package pulumitest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	res := test.Import("random:index/randomString:RandomString", "str", "importedString", "")

	// Assert on the generated YAML containing a resource definition
	require.Contains(t, res.Stdout, "type: random:RandomString")

	// Assert on the stack containing the resource state
	stack := test.ExportStack()
	data, err := stack.Deployment.MarshalJSON()
	require.NoError(t, err)
	var stateMap map[string]interface{}
	err = json.Unmarshal(data, &stateMap)
	require.NoError(t, err)

	resourcesJSON := stateMap["resources"].([]interface{})

	for _, res := range resourcesJSON {
		// get the id

		id := res.(map[string]interface{})["id"]
		if id == "importedString" {
			return
		}
	}

	t.Fatalf("resource not found in state")
}

func TestImportWithArgs(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_empty")

	outFile := filepath.Join(test.CurrentStack().Workspace().WorkDir(), "out.yaml")
	res := test.Import("random:index/randomString:RandomString", "str", "importedString", "", "--out", outFile)

	assert.Equal(t, []string{
		"import",
		"random:index/randomString:RandomString",
		"str", "importedString", "--yes", "--protect=false",
		"-s", test.CurrentStack().Name(), "--out", outFile,
	}, res.Args)

	// Assert on the generated YAML containing a resource definition
	contents, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "type: random:RandomString")

	// Assert on the stack containing the resource state
	stack := test.ExportStack()
	data, err := stack.Deployment.MarshalJSON()
	require.NoError(t, err)
	var stateMap map[string]interface{}
	err = json.Unmarshal(data, &stateMap)
	require.NoError(t, err)

	resourcesJSON := stateMap["resources"].([]interface{})

	for _, res := range resourcesJSON {
		// get the id

		id := res.(map[string]interface{})["id"]
		if id == "importedString" {
			return
		}
	}

	t.Fatalf("resource not found in state")
}
