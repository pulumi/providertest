package pulumitest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

var fixedPythonRuntime = `
runtime:
  name: python
  options:
    toolchain: pip
    virtualenv: venv
`

func TestImmediateConvertTypescript(t *testing.T) {
	t.Parallel()

	// No need to copy the source, since we're not going to modify it.
	convertResult := pulumitest.Convert(t, filepath.Join("testdata", "yaml_program"), "typescript")
	t.Log(convertResult.Output)
	test := convertResult.PulumiTest

	tsPreview := test.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		tsPreview.ChangeSummary)

	tsUp := test.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*tsUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, tsUp)

	// Show the deploy output.
	t.Log(tsUp.StdOut)
}

func TestConvertPython(t *testing.T) {
	t.Parallel()

	// No need to copy the source, since we're not going to modify it.
	source := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to Python.
	converted := source.Convert(t, "python", opttest.SkipInstall()).PulumiTest
	assert.NotEqual(t, converted.Source(), source.Source())

	// Fix up the python runtime to use venv.
	config, err := os.ReadFile(filepath.Join(converted.WorkingDir(), "Pulumi.yaml"))
	assert.NoError(t, err)
	config = []byte(strings.Replace(string(config), "runtime: python", fixedPythonRuntime, 1))
	err = os.WriteFile(filepath.Join(converted.WorkingDir(), "Pulumi.yaml"), config, 0644)
	assert.NoError(t, err)

	converted.Install(t)

	pythonPreview := converted.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		pythonPreview.ChangeSummary)

	pythonUp := converted.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*pythonUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, pythonUp)

	// Show the deploy output.
	t.Log(pythonUp.StdOut)
}

func TestConvertGo(t *testing.T) {
	t.Parallel()

	// No need to copy the source, since we're not going to modify it.
	source := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to Go.
	converted := source.Convert(t, "go").PulumiTest
	assert.NotEqual(t, converted.Source(), source.Source())

	goPreview := converted.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		goPreview.ChangeSummary)

	goUp := converted.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*goUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, goUp)

	// Show the deploy output.
	t.Log(goUp.StdOut)
}

func TestConvertTypescript(t *testing.T) {
	t.Parallel()

	// No need to copy the source, since we're not going to modify it.
	source := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to TypeScript.
	converted := source.Convert(t, "typescript").PulumiTest
	assert.NotEqual(t, converted.Source(), source.Source())

	tsPreview := converted.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		tsPreview.ChangeSummary)

	tsUp := converted.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*tsUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, tsUp)

	// Show the deploy output.
	t.Log(tsUp.StdOut)
}

func TestConvertCsharp(t *testing.T) {
	t.Parallel()

	// No need to copy the source, since we're not going to modify it.
	source := pulumitest.NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to C#.
	converted := source.Convert(t, "csharp").PulumiTest
	assert.NotEqual(t, converted.Source(), source.Source())

	csharpPreview := converted.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		csharpPreview.ChangeSummary)

	csharpUp := converted.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*csharpUp.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, csharpUp)

	// Show the deploy output.
	t.Log(csharpUp.StdOut)
}
