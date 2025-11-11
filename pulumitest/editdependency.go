package pulumitest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

// detectLanguage identifies the language/build system from the project directory
func detectLanguage(workingDir string) string {
	// Check for Go
	if _, err := os.Stat(filepath.Join(workingDir, "go.mod")); err == nil {
		return "go"
	}

	// Check for .NET
	if _, err := os.Stat(filepath.Join(workingDir, "Pulumi.yaml")); err == nil {
		if hasCsprojFile(workingDir) {
			return "dotnet"
		}
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(workingDir, "requirements.txt")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(workingDir, "setup.py")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(workingDir, "Pipfile")); err == nil {
		return "python"
	}

	// Check for Node.js
	if _, err := os.Stat(filepath.Join(workingDir, "package.json")); err == nil {
		return "nodejs"
	}

	// Check for YAML-only project (no specific language detected)
	if _, err := os.Stat(filepath.Join(workingDir, "Pulumi.yaml")); err == nil {
		return "yaml"
	}

	return "unknown"
}

// hasCsprojFile checks if there's a .csproj file in the directory or subdirectories
func hasCsprojFile(workingDir string) bool {
	entries, err := os.ReadDir(workingDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".csproj") {
			return true
		}
	}
	return false
}

// editDependencyNode edits a dependency in package.json (Node.js/npm/yarn)
func editDependencyNode(workingDir string, packageName string, version string) error {
	packageJSONPath := filepath.Join(workingDir, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Try to find and update the dependency in dependencies, devDependencies, or peerDependencies
	for _, depType := range []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"} {
		if deps, ok := packageJSON[depType].(map[string]interface{}); ok {
			if _, exists := deps[packageName]; exists {
				deps[packageName] = version
				data, err = json.MarshalIndent(packageJSON, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal package.json: %w", err)
				}
				data = append(data, '\n')
				if err := os.WriteFile(packageJSONPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write package.json: %w", err)
				}
				return nil
			}
		}
	}

	// If not found in existing dependencies, add to dependencies
	if deps, ok := packageJSON["dependencies"].(map[string]interface{}); ok {
		deps[packageName] = version
	} else {
		packageJSON["dependencies"] = map[string]interface{}{packageName: version}
	}

	data, err = json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(packageJSONPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}
	return nil
}

// editDependencyPython edits a dependency in Python files (pip/requirements.txt)
func editDependencyPython(workingDir string, packageName string, version string) error {
	// Try requirements.txt first
	requirementsPath := filepath.Join(workingDir, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		return updateRequirementsTxt(requirementsPath, packageName, version)
	}

	// Try setup.py if requirements.txt doesn't exist
	setupPyPath := filepath.Join(workingDir, "setup.py")
	if _, err := os.Stat(setupPyPath); err == nil {
		return fmt.Errorf("editing setup.py is not yet supported; use requirements.txt or pip install")
	}

	// Try Pipfile
	pipfilePath := filepath.Join(workingDir, "Pipfile")
	if _, err := os.Stat(pipfilePath); err == nil {
		return fmt.Errorf("editing Pipfile is not yet supported; use requirements.txt instead")
	}

	return fmt.Errorf("no Python dependency file found (requirements.txt, setup.py, or Pipfile)")
}

// updateRequirementsTxt updates a package version in requirements.txt
func updateRequirementsTxt(path string, packageName string, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read requirements.txt: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check if this line is for our package
		parts := strings.FieldsFunc(line, func(r rune) bool { return r == '=' || r == '<' || r == '>' || r == '!' || r == '~' })
		if len(parts) > 0 && strings.EqualFold(strings.TrimSpace(parts[0]), packageName) {
			lines[i] = fmt.Sprintf("%s==%s", packageName, version)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s==%s", packageName, version))
	}

	output := strings.Join(lines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write requirements.txt: %w", err)
	}
	return nil
}

// editDependencyGo edits a dependency in go.mod
func editDependencyGo(workingDir string, packageName string, version string) error {
	cmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", packageName, version))
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update go.mod: %v\n%s", err, out)
	}
	return nil
}

// editDependencyDotNet edits a dependency in .csproj files
func editDependencyDotNet(workingDir string, packageName string, version string) error {
	entries, err := os.ReadDir(workingDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var found bool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".csproj") {
			continue
		}

		csprojPath := filepath.Join(workingDir, entry.Name())
		if err := updateCsprojFile(csprojPath, packageName, version); err != nil {
			return err
		}
		found = true
	}

	if !found {
		return fmt.Errorf("no .csproj file found in directory")
	}
	return nil
}

// updateCsprojFile updates package version in a .csproj file
func updateCsprojFile(path string, packageName string, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	content := string(data)
	oldPattern := fmt.Sprintf("<PackageReference Include=\"%s\"", packageName)
	if !strings.Contains(content, oldPattern) {
		// Package not found, add it
		newPackageRef := fmt.Sprintf("\n    <PackageReference Include=\"%s\" Version=\"%s\" />", packageName, version)
		// Insert before closing </ItemGroup>
		content = strings.Replace(content, "</ItemGroup>", newPackageRef+"\n  </ItemGroup>", 1)
	} else {
		// Update existing package
		// Find and replace the version attribute
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(line, oldPattern) {
				// Extract indentation and replace the line
				indent := strings.Index(line, "<")
				lines[i] = strings.Repeat(" ", indent) + fmt.Sprintf("<PackageReference Include=\"%s\" Version=\"%s\" />", packageName, version)
				break
			}
		}
		content = strings.Join(lines, "\n")
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// editDependencies applies all dependency edits for the given language
func editDependencies(t PT, workingDir string, edits []opttest.DependencyEdit) error {
	if len(edits) == 0 {
		return nil
	}

	language := detectLanguage(workingDir)
	ptLogF(t, "detected language: %s", language)

	for _, edit := range edits {
		ptLogF(t, "editing dependency: %s=%s (language: %s)", edit.PackageName, edit.Version, language)
		switch language {
		case "nodejs":
			if err := editDependencyNode(workingDir, edit.PackageName, edit.Version); err != nil {
				return fmt.Errorf("failed to edit Node.js dependency: %w", err)
			}
		case "python":
			if err := editDependencyPython(workingDir, edit.PackageName, edit.Version); err != nil {
				return fmt.Errorf("failed to edit Python dependency: %w", err)
			}
		case "go":
			if err := editDependencyGo(workingDir, edit.PackageName, edit.Version); err != nil {
				return fmt.Errorf("failed to edit Go dependency: %w", err)
			}
		case "dotnet":
			if err := editDependencyDotNet(workingDir, edit.PackageName, edit.Version); err != nil {
				return fmt.Errorf("failed to edit .NET dependency: %w", err)
			}
		case "yaml":
			// For YAML projects, we can't automatically edit dependencies
			// Return a helpful error message
			return fmt.Errorf("cannot automatically edit dependencies in YAML-only projects; please configure manually or use a provider-specific mechanism")
		default:
			return fmt.Errorf("unknown language detected: %s", language)
		}
	}

	return nil
}
