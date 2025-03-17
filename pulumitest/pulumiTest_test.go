package pulumitest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestDeploy(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.SkipInstall(), opttest.SkipStackCreate())

	// Ensure dependencies are installed.
	test.Install(t)
	// Create a new stack with auto-naming.
	test.NewStack(t, "")

	// Test a preview.
	yamlPreview := test.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		yamlPreview.ChangeSummary)
	// Now do a real deploy.
	yamlUp := test.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*yamlUp.Summary.ResourceChanges)

	// Export the stack state.
	yamlStack := test.ExportStack(t)
	test.ImportStack(t, yamlStack)

	yamlPreview2 := test.Preview(t)
	assertpreview.HasNoChanges(t, yamlPreview2)
}

func TestConvert(t *testing.T) {
	t.Parallel()
	// No need to copy the source, since we're not going to modify it.
	source := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.TestInPlace())

	// Convert the original source to Python.
	converted := source.Convert(t, "python").PulumiTest
	assert.NotEqual(t, converted.Source(), source.Source())

	converted.Install(t)
	converted.NewStack(t, "test")

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

func TestGrpcLog(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"))
	test.Preview(t)
	grpcLog := test.GrpcLog(t)
	creates, err := grpcLog.Creates()
	assert.NoError(t, err, "expected no error when reading creates from grpc log")
	assert.Equal(t, 1, len(creates))
}

func TestDefaults(t *testing.T) {
	t.Parallel()
	source := filepath.Join("testdata", "yaml_program")
	test := NewPulumiTest(t, source)
	assert.NotEqual(t, source, test.Source(), "should copy source to a temporary directory")
	assert.NotNil(t, test.CurrentStack(), "should create a stack")
	assert.Equal(t, "test", test.CurrentStack().Name(), "should create a stack named 'test'")
	env := test.CurrentStack().Workspace().GetEnvVars()
	t.Log(env)
	assert.Equal(t, "correct horse battery staple", env["PULUMI_CONFIG_PASSPHRASE"], "should configure passphrase for encryption")
	assert.NotEmpty(t, env["PULUMI_BACKEND_URL"], "should configure backend URL")
	assert.NotEmpty(t, env["PULUMI_DEBUG_GRPC"], "should configure gRPC debug log")
	assert.Len(t, env, 3, "should not configure additional environment variables")
}

func TestInPlace(t *testing.T) {
	t.Parallel()
	source := filepath.Join("testdata", "yaml_program")
	test := NewPulumiTest(t, source, opttest.TestInPlace())
	assert.Equal(t, source, test.Source(), "should not copy source to a temporary directory")
	assert.NotNil(t, test.CurrentStack(), "should create a stack")
	assert.Equal(t, "test", test.CurrentStack().Name(), "should create a stack named 'test'")
	env := test.CurrentStack().Workspace().GetEnvVars()
	assert.Equal(t, "correct horse battery staple", env["PULUMI_CONFIG_PASSPHRASE"], "should configure passphrase for encryption")
	assert.NotEmpty(t, env["PULUMI_BACKEND_URL"], "should configure backend URL")
	assert.NotEmpty(t, env["PULUMI_DEBUG_GRPC"], "should configure gRPC debug log")
	assert.Len(t, env, 3, "should not configure additional environment variables")
}

func TestSkipStackCreate(t *testing.T) {
	t.Parallel()
	source := filepath.Join("testdata", "yaml_program")
	test := NewPulumiTest(t, source, opttest.SkipStackCreate())
	assert.NotEqual(t, source, test.Source(), "should copy source to a temporary directory")
	assert.Nil(t, test.CurrentStack(), "should not create a stack")
}

func TestSkipStackCreateInPlace(t *testing.T) {
	t.Parallel()
	source := filepath.Join("testdata", "yaml_program")
	test := NewPulumiTest(t, source, opttest.SkipStackCreate(), opttest.TestInPlace())
	assert.Equal(t, source, test.Source(), "should not copy source to a temporary directory")
	assert.Nil(t, test.CurrentStack(), "should not create a stack")
}

func TestProviderPluginPath(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "yaml_program"), opttest.DownloadProviderVersion("random", "4.15.0"))
	test.Preview(t)

	settings, err := test.CurrentStack().Workspace().ProjectSettings(test.Context())
	assert.NoError(t, err, "expected no error when getting project settings")
	snaps.MatchJSON(t, settings.Plugins.Providers, match.Any("0.path"))
}

func TestCustomTempDir(t *testing.T) {
	t.Parallel()
	// Ensure ".temp" doesn't yet exist.
	assert.NoError(t, os.RemoveAll(".temp"))
	// Test installing python program in a custom local directory.
	NewPulumiTest(t, filepath.Join("testdata", "python_gcp"), opttest.TempDir(".temp/sub-dir"))

	assert.DirExists(t, ".temp", "should leave custom local .temp directory")
}
