package autotest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStackPreview(t *testing.T) {
	sourceTest := NewAutoTest(t, filepath.Join("testdata", "yaml_program"))

	yamlTest := sourceTest.CopyToTempDir()
	yamlTest.Env().UseLocalBackend()
	yamlTest.Install()
	stackName := "stack"
	yamlTest.NewStack(stackName)
	yamlPreview := yamlTest.Preview()
	yamlUp := yamlTest.Up()
	assert.Equal(t, 1, len(*yamlUp.Summary.ResourceChanges))

	yamlStack := yamlTest.ExportStack()

	pythonConvert := sourceTest.Convert("python")
	pythonTest := pythonConvert.AutoTest
	pythonTest.Env().UseLocalBackend()
	pythonTest.Install()
	pythonTest.NewStack(stackName)
	pythonTest.ImportStack(yamlStack)

	pythonPreview := pythonTest.Preview()
	assert.Equal(t, yamlPreview.ChangeSummary, pythonPreview.ChangeSummary)

	pythonUp := pythonTest.Up()
	assert.Equal(t, 0, len(*pythonUp.Summary.ResourceChanges))
}
