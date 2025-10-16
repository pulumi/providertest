package pulumitest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectLanguageGo verifies Go project detection
func TestDetectLanguageGo(t *testing.T) {
	tmpDir := t.TempDir()
	gomodPath := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(gomodPath, []byte("module example.com/test\n"), 0644)
	require.NoError(t, err)

	lang := detectLanguage(tmpDir)
	assert.Equal(t, "go", lang)
}

// TestDetectLanguageNodeJS verifies Node.js project detection
func TestDetectLanguageNodeJS(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSONPath := filepath.Join(tmpDir, "package.json")
	err := os.WriteFile(packageJSONPath, []byte(`{"name": "test"}`), 0644)
	require.NoError(t, err)

	lang := detectLanguage(tmpDir)
	assert.Equal(t, "nodejs", lang)
}

// TestDetectLanguagePython verifies Python project detection with requirements.txt
func TestDetectLanguagePython(t *testing.T) {
	tmpDir := t.TempDir()
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	err := os.WriteFile(reqPath, []byte("requests==2.28.0\n"), 0644)
	require.NoError(t, err)

	lang := detectLanguage(tmpDir)
	assert.Equal(t, "python", lang)
}

// TestDetectLanguageDotNet verifies .NET project detection
func TestDetectLanguageDotNet(t *testing.T) {
	tmpDir := t.TempDir()
	pulumiYamlPath := filepath.Join(tmpDir, "Pulumi.yaml")
	err := os.WriteFile(pulumiYamlPath, []byte("name: test\n"), 0644)
	require.NoError(t, err)

	csprojPath := filepath.Join(tmpDir, "test.csproj")
	err = os.WriteFile(csprojPath, []byte("<Project></Project>"), 0644)
	require.NoError(t, err)

	lang := detectLanguage(tmpDir)
	assert.Equal(t, "dotnet", lang)
}

// TestEditDependencyNode verifies Node.js dependency editing
func TestEditDependencyNode(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSONPath := filepath.Join(tmpDir, "package.json")

	// Create initial package.json
	packageJSON := map[string]interface{}{
		"name": "test",
		"dependencies": map[string]interface{}{
			"lodash": "^4.17.0",
		},
	}
	data, err := json.MarshalIndent(packageJSON, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(packageJSONPath, append(data, '\n'), 0644)
	require.NoError(t, err)

	// Edit lodash version
	err = editDependencyNode(tmpDir, "lodash", "^4.18.0")
	require.NoError(t, err)

	// Verify the edit
	data, err = os.ReadFile(packageJSONPath)
	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	deps := result["dependencies"].(map[string]interface{})
	assert.Equal(t, "^4.18.0", deps["lodash"])
}

// TestEditDependencyNodeAdd verifies adding a new dependency to Node.js
func TestEditDependencyNodeAdd(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSONPath := filepath.Join(tmpDir, "package.json")

	// Create initial package.json without the dependency
	packageJSON := map[string]interface{}{
		"name": "test",
	}
	data, err := json.MarshalIndent(packageJSON, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(packageJSONPath, append(data, '\n'), 0644)
	require.NoError(t, err)

	// Add new dependency
	err = editDependencyNode(tmpDir, "express", "^4.18.0")
	require.NoError(t, err)

	// Verify the addition
	data, err = os.ReadFile(packageJSONPath)
	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	deps := result["dependencies"].(map[string]interface{})
	assert.Equal(t, "^4.18.0", deps["express"])
}

// TestUpdateRequirementsTxt verifies Python requirements.txt editing
func TestUpdateRequirementsTxt(t *testing.T) {
	tmpDir := t.TempDir()
	reqPath := filepath.Join(tmpDir, "requirements.txt")

	// Create initial requirements.txt
	initialContent := "requests==2.28.0\nnumpy==1.21.0\n"
	err := os.WriteFile(reqPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Update requests version
	err = updateRequirementsTxt(reqPath, "requests", "2.31.0")
	require.NoError(t, err)

	// Verify the edit
	data, err := os.ReadFile(reqPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "requests==2.31.0")
	assert.Contains(t, content, "numpy==1.21.0")
}

// TestUpdateRequirementsTxtAdd verifies adding a new package to requirements.txt
func TestUpdateRequirementsTxtAdd(t *testing.T) {
	tmpDir := t.TempDir()
	reqPath := filepath.Join(tmpDir, "requirements.txt")

	// Create initial requirements.txt
	initialContent := "requests==2.28.0\n"
	err := os.WriteFile(reqPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Add new package
	err = updateRequirementsTxt(reqPath, "django", "4.1.0")
	require.NoError(t, err)

	// Verify the addition
	data, err := os.ReadFile(reqPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "django==4.1.0")
	assert.Contains(t, content, "requests==2.28.0")
}

// TestHasCsprojFile verifies .csproj file detection
func TestHasCsprojFile(t *testing.T) {
	tmpDir := t.TempDir()
	csprojPath := filepath.Join(tmpDir, "test.csproj")
	err := os.WriteFile(csprojPath, []byte("<Project></Project>"), 0644)
	require.NoError(t, err)

	has := hasCsprojFile(tmpDir)
	assert.True(t, has)
}

// TestHasCsprojFileNoMatch verifies hasCsprojFile returns false when no .csproj exists
func TestHasCsprojFileNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	has := hasCsprojFile(tmpDir)
	assert.False(t, has)
}

// TestUpdateCsprojFileNew verifies adding a new package reference to .csproj
func TestUpdateCsprojFileNew(t *testing.T) {
	tmpDir := t.TempDir()
	csprojPath := filepath.Join(tmpDir, "test.csproj")

	// Create initial .csproj with ItemGroup
	csprojContent := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
  <ItemGroup>
  </ItemGroup>
</Project>`
	err := os.WriteFile(csprojPath, []byte(csprojContent), 0644)
	require.NoError(t, err)

	// Add new package reference
	err = updateCsprojFile(csprojPath, "Pulumi", "3.0.0")
	require.NoError(t, err)

	// Verify the addition
	data, err := os.ReadFile(csprojPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `<PackageReference Include="Pulumi" Version="3.0.0" />`)
}

// TestUpdateCsprojFileExisting verifies updating an existing package reference
func TestUpdateCsprojFileExisting(t *testing.T) {
	tmpDir := t.TempDir()
	csprojPath := filepath.Join(tmpDir, "test.csproj")

	// Create .csproj with existing package reference
	csprojContent := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Pulumi" Version="2.0.0" />
  </ItemGroup>
</Project>`
	err := os.WriteFile(csprojPath, []byte(csprojContent), 0644)
	require.NoError(t, err)

	// Update package version
	err = updateCsprojFile(csprojPath, "Pulumi", "3.0.0")
	require.NoError(t, err)

	// Verify the update
	data, err := os.ReadFile(csprojPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `<PackageReference Include="Pulumi" Version="3.0.0" />`)
	assert.NotContains(t, content, `Version="2.0.0"`)
}

// TestEditDependencyNodeDevDependencies verifies editing dev dependencies
func TestEditDependencyNodeDevDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSONPath := filepath.Join(tmpDir, "package.json")

	// Create package.json with devDependencies
	packageJSON := map[string]interface{}{
		"name": "test",
		"devDependencies": map[string]interface{}{
			"jest": "^27.0.0",
		},
	}
	data, err := json.MarshalIndent(packageJSON, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(packageJSONPath, append(data, '\n'), 0644)
	require.NoError(t, err)

	// Update jest version in devDependencies
	err = editDependencyNode(tmpDir, "jest", "^29.0.0")
	require.NoError(t, err)

	// Verify the edit
	data, err = os.ReadFile(packageJSONPath)
	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	devDeps := result["devDependencies"].(map[string]interface{})
	assert.Equal(t, "^29.0.0", devDeps["jest"])
}

// TestEditDependencyPythonNoFile verifies error when no Python file exists
func TestEditDependencyPythonNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to edit dependency without any Python file
	err := editDependencyPython(tmpDir, "requests", "2.31.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Python dependency file found")
}

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
	content := string(data)
	assert.Contains(t, content, "v3.100.0")
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

func TestEditDependencyDotNet(t *testing.T) {
	tmpDir := t.TempDir()
	csprojPath := filepath.Join(tmpDir, "test.csproj")

	csprojContent := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Pulumi" Version="3.0.0" />
  </ItemGroup>
</Project>`
	err := os.WriteFile(csprojPath, []byte(csprojContent), 0644)
	require.NoError(t, err)

	err = editDependencyDotNet(tmpDir, "Pulumi", "3.50.0")
	require.NoError(t, err)

	data, err := os.ReadFile(csprojPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `Version="3.50.0"`)
}

func TestEditDependencyDotNetNoCsproj(t *testing.T) {
	tmpDir := t.TempDir()

	err := editDependencyDotNet(tmpDir, "Pulumi", "3.50.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no .csproj file found")
}

func TestEditDependenciesNodeJS(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := map[string]interface{}{
		"name": "test",
		"dependencies": map[string]interface{}{
			"lodash": "^4.17.0",
		},
	}
	data, _ := json.MarshalIndent(packageJSON, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), append(data, '\n'), 0644))

	edits := []opttest.DependencyEdit{
		{PackageName: "lodash", Version: "^4.18.0"},
	}

	err := editDependencies(t, tmpDir, edits)
	require.NoError(t, err)

	data, _ = os.ReadFile(filepath.Join(tmpDir, "package.json"))
	assert.Contains(t, string(data), `"lodash": "^4.18.0"`)
}

func TestEditDependenciesYAMLError(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Pulumi.yaml"), []byte("name: test\n"), 0644))

	edits := []opttest.DependencyEdit{
		{PackageName: "pulumi", Version: "3.50.0"},
	}

	err := editDependencies(t, tmpDir, edits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YAML-only projects")
}

func TestEditDependenciesUnknownLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	edits := []opttest.DependencyEdit{
		{PackageName: "some-package", Version: "1.0.0"},
	}

	err := editDependencies(t, tmpDir, edits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown language")
}

func TestUpdateRequirementsTxtComplexVersions(t *testing.T) {
	tmpDir := t.TempDir()
	reqPath := filepath.Join(tmpDir, "requirements.txt")

	initialContent := `requests==2.28.0
pulumi-aws==5.0.0
numpy==1.21.0
`
	err := os.WriteFile(reqPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		pkg     string
		version string
	}{
		{"requests", "2.31.0"},
		{"pulumi-aws", "6.0.0"},
		{"numpy", "1.24.0"},
	}

	for _, tt := range tests {
		err = updateRequirementsTxt(reqPath, tt.pkg, tt.version)
		require.NoError(t, err)

		data, _ := os.ReadFile(reqPath)
		content := string(data)
		assert.Contains(t, content, fmt.Sprintf("%s==%s", tt.pkg, tt.version))
	}
}

func TestDetectLanguagePriority(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test"}`), 0644))

	lang := detectLanguage(tmpDir)
	assert.Equal(t, "go", lang)
}
