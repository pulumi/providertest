package pulumitest

import (
	"os"
	"path/filepath"
	"strings"
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
	customTempDir := t.TempDir()
	// Test installing python program in a custom local directory.
	test := NewPulumiTest(t, filepath.Join("testdata", "python_gcp"), opttest.TempDir(customTempDir))

	workingDir := test.WorkingDir()
	if !strings.HasPrefix(workingDir, customTempDir) {
		t.Fatalf("expected working directory to be in the custom temp directory, got %s", workingDir)
	}
}

func TestDotNetDeploy(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "csharp_simple"))

	// Test a preview.
	preview := test.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		preview.ChangeSummary)

	// Now do a real deploy.
	up := test.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*up.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, up)

	// Verify outputs exist
	assert.NotEmpty(t, up.Outputs["name"].Value)

	// Test that a second preview shows no changes
	preview2 := test.Preview(t)
	assertpreview.HasNoChanges(t, preview2)
}

func TestDotNetSkipInstall(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "csharp_simple"), opttest.SkipInstall(), opttest.SkipStackCreate())

	// Manually install and create stack
	test.Install(t)
	test.NewStack(t, "test")

	// Test a preview.
	preview := test.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		preview.ChangeSummary)
}

func TestDotNetWithLocalReference(t *testing.T) {
	t.Parallel()
	mockSdkPath := filepath.Join("testdata", "mock_sdk")

	// Create test with local project reference to mock SDK
	// Skip install since we don't need to build, just verify .csproj modification
	test := NewPulumiTest(t,
		filepath.Join("testdata", "csharp_with_ref"),
		opttest.DotNetReference("MockSdk", mockSdkPath),
		opttest.SkipInstall())

	// Verify that the .csproj was modified with the reference
	// The modification happens during NewStack which is called automatically
	csprojPath := filepath.Join(test.WorkingDir(), "csharp_with_ref.csproj")
	csprojContent, err := os.ReadFile(csprojPath)
	assert.NoError(t, err, "should be able to read .csproj")
	assert.Contains(t, string(csprojContent), "<ProjectReference", "should contain ProjectReference")
	assert.Contains(t, string(csprojContent), "MockSdk.csproj", "should reference MockSdk.csproj")

	t.Log(".csproj successfully modified with local project reference")
}

func TestJavaMavenTargetVersion(t *testing.T) {
	t.Parallel()
	// Test that JavaTargetVersion option can be created without errors
	opts := opttest.DefaultOptions()
	opttest.JavaTargetVersion("17").Apply(opts)

	// Verify the option was applied
	assert.Equal(t, "17", opts.JavaTargetVersion,
		"should set Java target version to 17")
}

func TestJavaMavenProfile(t *testing.T) {
	t.Parallel()
	// Test that JavaMavenProfile option can be created without errors
	opts := opttest.DefaultOptions()
	opttest.JavaMavenProfile("development").Apply(opts)

	// Verify the option was applied
	assert.Equal(t, "development", opts.JavaMavenProfile,
		"should set Maven profile to development")
}

func TestJavaMavenSettings(t *testing.T) {
	t.Parallel()
	// Create a temporary settings file
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.xml")
	err := os.WriteFile(settingsPath, []byte("<settings></settings>"), 0644)
	assert.NoError(t, err, "should create settings file")

	// Test that JavaMavenSettings option can be created without errors
	opts := opttest.DefaultOptions()
	opttest.JavaMavenSettings(settingsPath).Apply(opts)

	// Verify the option was applied
	assert.Equal(t, settingsPath, opts.JavaMavenSettings,
		"should set Maven settings path")
}

func TestJavaMavenDependency(t *testing.T) {
	t.Parallel()
	// Create a test directory to use as local Maven dependency
	depDir := t.TempDir()

	// Test that JavaMavenDependency option can be created without errors
	opts := opttest.DefaultOptions()
	opttest.JavaMavenDependency("com.example", "mylib", depDir).Apply(opts)

	// Verify the option was applied
	assert.NotEmpty(t, opts.JavaMavenDependencies,
		"should add Maven dependency")

	key := "com.example:mylib"
	assert.Contains(t, opts.JavaMavenDependencies, key,
		"should have dependency key")

	dep := opts.JavaMavenDependencies[key]
	assert.Equal(t, "com.example", dep.GroupID, "should set groupId")
	assert.Equal(t, "mylib", dep.ArtifactID, "should set artifactId")
	assert.Equal(t, depDir, dep.Path, "should set path")
}
