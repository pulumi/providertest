package autotest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStackPreview(t *testing.T) {
	sourceTest := NewAutoTest(t, filepath.Join("testdata", "yaml_program"))

	yamlTest := sourceTest.CopyToTempDir()
	yamlTest.Install()
	yamlStack := yamlTest.NewStack("")
	yamlPreview := yamlTest.Preview(yamlStack)

	pythonConvert := sourceTest.Convert("python")
	pythonTest := pythonConvert.AutoTest
	pythonTest.Install()
	pythonStack := pythonTest.NewStack("")
	pythonPreview := pythonTest.Preview(pythonStack)

	assert.Equal(t, yamlPreview.ChangeSummary, pythonPreview.ChangeSummary)

	yamlUp := yamlTest.Up(yamlStack)
	pythonUp := pythonTest.Up(pythonStack)

	assert.Equal(t, yamlUp.Summary.ResourceChanges, pythonUp.Summary.ResourceChanges)
}
