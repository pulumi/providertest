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
	"github.com/stretchr/testify/require"
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

// TestJavaMavenWithLocalReference is an integration test that mirrors
// TestDotNetWithLocalReference. It verifies that Java/Maven options actually
// modify pom.xml during NewStack, not just populate the options struct.
// SkipInstall is used so no Maven toolchain is required.
func TestJavaMavenWithLocalReference(t *testing.T) {
	t.Parallel()

	// Create a fake local artifact directory used as a system-scoped dependency.
	depDir := t.TempDir()
	jarPath := filepath.Join(depDir, "mylib-0.0.0-dev.jar")
	err := os.WriteFile(jarPath, []byte("fake jar content"), 0644)
	require.NoError(t, err)

	test := NewPulumiTest(t,
		filepath.Join("testdata", "java_simple"),
		opttest.JavaTargetVersion("17"),
		opttest.JavaMavenDependency("com.example", "mylib", depDir),
		opttest.SkipInstall(),
	)

	pomPath := filepath.Join(test.WorkingDir(), "pom.xml")
	pomContent, err := os.ReadFile(pomPath)
	assert.NoError(t, err, "should be able to read pom.xml")

	pomStr := string(pomContent)
	// Match content + closing tag only; the encoder rewrites opening tags with xmlns attributes.
	assert.Contains(t, pomStr, ">17</maven.compiler.source>",
		"should update maven.compiler.source to 17")
	assert.Contains(t, pomStr, ">17</maven.compiler.target>",
		"should update maven.compiler.target to 17")
	assert.Contains(t, pomStr, ">17</source>",
		"should update compiler plugin source to 17")
	assert.Contains(t, pomStr, ">17</target>",
		"should update compiler plugin target to 17")
	assert.Contains(t, pomStr, ">com.example</groupId>",
		"should add Maven dependency groupId")
	assert.Contains(t, pomStr, ">mylib</artifactId>",
		"should add Maven dependency artifactId")
	assert.Contains(t, pomStr, ">system</scope>",
		"local Maven dependency should use system scope")
	assert.Contains(t, pomStr, jarPath,
		"directory dependency should resolve to a concrete JAR path")

	t.Log("pom.xml successfully modified with Java target version and local Maven dependency")
}

func TestJavaMavenEnvOptionsReachWorkspace(t *testing.T) {
	t.Parallel()

	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	err := os.WriteFile(settingsPath, []byte("<settings></settings>"), 0644)
	require.NoError(t, err)

	test := NewPulumiTest(t,
		filepath.Join("testdata", "java_simple"),
		opttest.JavaMavenProfile("development"),
		opttest.JavaMavenSettings(settingsPath),
		opttest.SkipInstall(),
	)

	env := test.CurrentStack().Workspace().GetEnvVars()
	assert.Equal(t, "development", env["MAVEN_ACTIVE_PROFILES"])
	assert.Equal(t, settingsPath, env["MAVEN_SETTINGS"])
}

func TestJavaMavenEnvOptionsDoNotRequirePom(t *testing.T) {
	t.Parallel()

	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	err := os.WriteFile(settingsPath, []byte("<settings></settings>"), 0644)
	require.NoError(t, err)

	test := NewPulumiTest(t,
		filepath.Join("testdata", "nodejs_empty"),
		opttest.JavaMavenProfile("development"),
		opttest.JavaMavenSettings(settingsPath),
		opttest.SkipInstall(),
	)

	env := test.CurrentStack().Workspace().GetEnvVars()
	assert.Equal(t, "development", env["MAVEN_ACTIVE_PROFILES"])
	assert.Equal(t, settingsPath, env["MAVEN_SETTINGS"])
}

func TestJavaMavenOptionsAppliedBeforeInstall(t *testing.T) {
	sourceDir := t.TempDir()
	for _, name := range []string{"Pulumi.yaml", "pom.xml"} {
		content, err := os.ReadFile(filepath.Join("testdata", "java_simple", name))
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(sourceDir, name), content, 0644)
		require.NoError(t, err)
	}

	depDir := t.TempDir()
	jarPath := filepath.Join(depDir, "mylib-0.0.0-dev.jar")
	err := os.WriteFile(jarPath, []byte("fake jar content"), 0644)
	require.NoError(t, err)

	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	err = os.WriteFile(settingsPath, []byte("<settings></settings>"), 0644)
	require.NoError(t, err)

	binDir := t.TempDir()
	pulumiPath := filepath.Join(binDir, "pulumi")
	pulumiScript := `#!/bin/sh
printf '%s' "$MAVEN_ACTIVE_PROFILES" > install-profile.txt
printf '%s' "$MAVEN_SETTINGS" > install-settings.txt
cp pom.xml install-pom.xml
`
	err = os.WriteFile(pulumiPath, []byte(pulumiScript), 0755)
	require.NoError(t, err)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	test := NewPulumiTest(t,
		sourceDir,
		opttest.TestInPlace(),
		opttest.SkipStackCreate(),
		opttest.JavaTargetVersion("17"),
		opttest.JavaMavenDependency("com.example", "mylib", depDir),
		opttest.JavaMavenProfile("development"),
		opttest.JavaMavenSettings(settingsPath),
	)

	installPom, err := os.ReadFile(filepath.Join(test.WorkingDir(), "install-pom.xml"))
	require.NoError(t, err)
	installPomStr := string(installPom)
	assert.Contains(t, installPomStr, ">17</maven.compiler.source>")
	assert.Contains(t, installPomStr, ">17</maven.compiler.target>")
	assert.Contains(t, installPomStr, ">com.example</groupId>")
	assert.Contains(t, installPomStr, jarPath)

	profile, err := os.ReadFile(filepath.Join(test.WorkingDir(), "install-profile.txt"))
	require.NoError(t, err)
	assert.Equal(t, "development", string(profile))

	installSettings, err := os.ReadFile(filepath.Join(test.WorkingDir(), "install-settings.txt"))
	require.NoError(t, err)
	assert.Equal(t, settingsPath, string(installSettings))
}

func TestResolveJavaMavenDependencyPathRejectsNonJarFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "not-a-jar.txt")
	err := os.WriteFile(filePath, []byte("not a jar"), 0644)
	require.NoError(t, err)

	_, err = resolveJavaMavenDependencyPath(filePath, "mylib")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must point to a .jar file")
}
