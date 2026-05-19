//go:build integration

package pulumitest

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	randomSDKUpgradeVersion = "4.16.0"

	randomPkgNode   = "@pulumi/random"
	randomPkgPy     = "pulumi-random"
	randomPkgGo     = "github.com/pulumi/pulumi-random/sdk/v4"
	randomPkgDotNet = "Pulumi.Random"
	randomPkgYAML   = "random"
)

func TestEditDependencyGo(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	initialContent := `module example.com/test

go 1.23

require (
	github.com/pulumi/pulumi/sdk/v3 v3.50.0
)
`
	err := os.WriteFile(goModPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	err = editDependencyGo(tmpDir, "github.com/pulumi/pulumi/sdk/v3", "v3.100.0")
	require.NoError(t, err)

	data, err := os.ReadFile(goModPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "v3.100.0")
}

func TestEditDependencyGoInvalidModule(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	err := os.WriteFile(goModPath, []byte("invalid content"), 0644)
	require.NoError(t, err)

	err = editDependencyGo(tmpDir, "github.com/example/package", "v1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update go.mod")
}

func TestEditDep_Node_FileEdit(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/nodejs_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgNode, randomSDKUpgradeVersion),
	)

	pkgJSON := readPackageJSON(t, filepath.Join(test.WorkingDir(), "package.json"))
	deps, _ := pkgJSON["dependencies"].(map[string]interface{})
	require.NotNil(t, deps, "dependencies missing from package.json")
	assert.Equal(t, randomSDKUpgradeVersion, deps[randomPkgNode])
}

func TestEditDep_Python_FileEdit(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/python_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgPy, randomSDKUpgradeVersion),
	)

	reqPath := filepath.Join(test.WorkingDir(), "requirements.txt")
	data, err := os.ReadFile(reqPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), randomPkgPy+"=="+randomSDKUpgradeVersion)
}

func TestEditDep_Go_FileEdit(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/go_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgGo, "v"+randomSDKUpgradeVersion),
	)

	data, err := os.ReadFile(filepath.Join(test.WorkingDir(), "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "v"+randomSDKUpgradeVersion)
}

func TestEditDep_DotNet_FileEdit(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/csharp_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgDotNet, randomSDKUpgradeVersion),
	)

	csproj, err := findCsprojFile(test.WorkingDir())
	require.NoError(t, err)
	data, err := os.ReadFile(csproj)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, randomPkgDotNet)
	assert.Contains(t, content, `Version="`+randomSDKUpgradeVersion+`"`)
}

func TestEditDep_YAML_FileEdit(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/yaml_random_explicit",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgYAML, randomSDKUpgradeVersion),
	)

	data, err := os.ReadFile(filepath.Join(test.WorkingDir(), "Pulumi.yaml"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "type: pulumi:providers:"+randomPkgYAML)
	assert.Contains(t, content, "version: "+randomSDKUpgradeVersion)
	// The stale baseline version from the testdata should no longer be present,
	// not just the new one appended.
	assert.NotContains(t, content, "version: 4.13.0")
}

func TestEditDep_Multi(t *testing.T) {
	t.Parallel()
	const otherPkg = "lodash"
	const otherVer = "4.17.21"

	test := NewPulumiTest(t, "testdata/nodejs_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgNode, randomSDKUpgradeVersion),
		opttest.EditDependency(otherPkg, otherVer),
	)

	pkgJSON := readPackageJSON(t, filepath.Join(test.WorkingDir(), "package.json"))
	deps, _ := pkgJSON["dependencies"].(map[string]interface{})
	require.NotNil(t, deps)
	assert.Equal(t, randomSDKUpgradeVersion, deps[randomPkgNode])
	assert.Equal(t, otherVer, deps[otherPkg])
}

func TestEditDep_SourceNotMutated(t *testing.T) {
	t.Parallel()
	sourcePath := filepath.Join("testdata", "nodejs_random", "package.json")
	before, err := os.ReadFile(sourcePath)
	require.NoError(t, err)

	NewPulumiTest(t, "testdata/nodejs_random",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgNode, randomSDKUpgradeVersion),
	)

	after, err := os.ReadFile(sourcePath)
	require.NoError(t, err)
	assert.Equal(t, before, after, "EditDependency must not mutate the source testdata file")
}

func TestEditDep_Node_Install(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/nodejs_random",
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgNode, randomSDKUpgradeVersion),
	)

	pkgJSON := readPackageJSON(t, filepath.Join(test.WorkingDir(), "package.json"))
	deps, _ := pkgJSON["dependencies"].(map[string]interface{})
	require.NotNil(t, deps)
	assert.Equal(t, randomSDKUpgradeVersion, deps[randomPkgNode])

	installed := readPackageJSON(t,
		filepath.Join(test.WorkingDir(), "node_modules", "@pulumi", "random", "package.json"))
	installedVer, _ := installed["version"].(string)
	// Published @pulumi/random packages have carried a v-prefix in package.json
	// at various points; strip it so the semver content is what's compared.
	assert.Equal(t, randomSDKUpgradeVersion, strings.TrimPrefix(installedVer, "v"))
}

func TestEditDep_Python_Install(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/python_random",
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgPy, randomSDKUpgradeVersion),
	)

	data, err := os.ReadFile(filepath.Join(test.WorkingDir(), "requirements.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(data), randomPkgPy+"=="+randomSDKUpgradeVersion)

	installed := findInstalledPythonPackageVersion(t, test.WorkingDir(), "pulumi_random")
	assert.Equal(t, randomSDKUpgradeVersion, installed)
}

func TestEditDep_Go_Install(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/go_random",
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgGo, "v"+randomSDKUpgradeVersion),
	)

	data, err := os.ReadFile(filepath.Join(test.WorkingDir(), "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "v"+randomSDKUpgradeVersion)

	installed := goListModVersion(t, test.WorkingDir(), randomPkgGo)
	assert.Equal(t, "v"+randomSDKUpgradeVersion, installed)
}

func TestEditDep_DotNet_Install(t *testing.T) {
	t.Parallel()
	test := NewPulumiTest(t, "testdata/csharp_random",
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgDotNet, randomSDKUpgradeVersion),
	)

	csproj, err := findCsprojFile(test.WorkingDir())
	require.NoError(t, err)
	data, err := os.ReadFile(csproj)
	require.NoError(t, err)
	assert.Contains(t, string(data), `Version="`+randomSDKUpgradeVersion+`"`)

	assertProjectAssetsContainsVersion(t,
		filepath.Join(test.WorkingDir(), "obj", "project.assets.json"),
		randomPkgDotNet, randomSDKUpgradeVersion)
}

func TestEditDep_UpDestroy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		dir     string
		pkg     string
		version string
	}{
		{"Node", "testdata/nodejs_random", randomPkgNode, randomSDKUpgradeVersion},
		{"Python", "testdata/python_random", randomPkgPy, randomSDKUpgradeVersion},
		{"Go", "testdata/go_random", randomPkgGo, "v" + randomSDKUpgradeVersion},
		{"DotNet", "testdata/csharp_random", randomPkgDotNet, randomSDKUpgradeVersion},
		{"YAML", "testdata/yaml_random_explicit", randomPkgYAML, randomSDKUpgradeVersion},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test := NewPulumiTest(t, tc.dir,
				opttest.EditDependency(tc.pkg, tc.version),
			)
			up := test.Up(t)
			assertNameOutput(t, up.Outputs)
		})
	}
}

func TestEditDep_WarnsOnLocalProvider(t *testing.T) {
	t.Parallel()
	ct := newCapturingT(t)

	NewPulumiTest(ct, "testdata/yaml_random_explicit",
		opttest.SkipInstall(),
		opttest.SkipStackCreate(),
		opttest.EditDependency(randomPkgYAML, randomSDKUpgradeVersion),
		opttest.LocalProviderPath("random", t.TempDir()),
	)

	assert.True(t, ct.sawLogContaining("WARNING: EditDependency is being used alongside a provider path"),
		"expected warning log when EditDependency and LocalProviderPath are combined; captured logs:\n%s",
		strings.Join(ct.logs(), "\n"))
}

func readPackageJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "reading %s", path)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &out), "parsing %s", path)
	return out
}

func goListModVersion(t *testing.T, workingDir, modulePath string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Version}}", modulePath)
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go list -m %s failed: %s", modulePath, out)
	return strings.TrimSpace(string(out))
}

func assertProjectAssetsContainsVersion(t *testing.T, path, packageName, version string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "reading %s", path)
	var parsed struct {
		Libraries map[string]json.RawMessage `json:"libraries"`
	}
	require.NoError(t, json.Unmarshal(data, &parsed))
	require.NotNil(t, parsed.Libraries, "project.assets.json missing 'libraries'")

	want := packageName + "/" + version
	for key := range parsed.Libraries {
		if strings.EqualFold(key, want) {
			return
		}
	}
	t.Fatalf("expected %q in project.assets.json libraries, got keys: %v",
		want, slices.Collect(maps.Keys(parsed.Libraries)))
}

// findInstalledPythonPackageVersion requires the program's Pulumi.yaml to set
// options.virtualenv: venv so pip installs land inside workingDir/venv rather
// than a host-global site-packages.
func findInstalledPythonPackageVersion(t *testing.T, workingDir, packageName string) string {
	t.Helper()
	sitePackagesGlob := filepath.Join(workingDir, "venv", "lib", "python*", "site-packages", packageName+"-*.dist-info", "METADATA")
	matches, err := filepath.Glob(sitePackagesGlob)
	require.NoError(t, err, "glob %s", sitePackagesGlob)
	require.NotEmpty(t, matches, "no METADATA for %s under %s", packageName, sitePackagesGlob)

	data, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Version: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version: "))
		}
	}
	t.Fatalf("no Version: line in %s", matches[0])
	return ""
}

func assertNameOutput(t *testing.T, outputs auto.OutputMap) {
	t.Helper()
	v, ok := outputs["name"]
	require.True(t, ok, "stack output 'name' missing; outputs=%v", outputs)
	s, _ := v.Value.(string)
	assert.NotEmpty(t, s, "stack output 'name' was empty")
}

type capturingT struct {
	*testing.T
	mu      sync.Mutex
	entries []string
}

func newCapturingT(t *testing.T) *capturingT {
	t.Helper()
	return &capturingT{T: t}
}

func (c *capturingT) Log(args ...any) {
	line := fmt.Sprint(args...)
	c.mu.Lock()
	c.entries = append(c.entries, line)
	c.mu.Unlock()
	c.T.Log(args...)
}

func (c *capturingT) logs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.entries...)
}

func (c *capturingT) sawLogContaining(needle string) bool {
	for _, e := range c.logs() {
		if strings.Contains(e, needle) {
			return true
		}
	}
	return false
}
