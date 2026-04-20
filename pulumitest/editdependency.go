package pulumitest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"gopkg.in/yaml.v3"
)

var requirementsLinePattern = regexp.MustCompile(`^(\s*)([^=<>!~;\s#]+(?:\[[^\]]+\])?)(\s*)(===|==|~=|!=|<=|>=|<|>)(\s*)([^;\s#]+)(.*)$`)

// detectLanguage identifies the language/build system from the project directory
func detectLanguage(workingDir string) string {
	// Check for Go
	if _, err := os.Stat(filepath.Join(workingDir, "go.mod")); err == nil {
		return "go"
	}

	// Check for .NET (.csproj presence is sufficient — no Pulumi.yaml guard needed)
	if hasCsprojFile(workingDir) {
		return "dotnet"
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

// hasCsprojFile checks if there's a .csproj file in the given directory (non-recursive).
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
		trimmedLine := strings.TrimSpace(line)
		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		matches := requirementsLinePattern.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		rawName := matches[2]
		if idx := strings.IndexByte(rawName, '['); idx >= 0 {
			rawName = rawName[:idx]
		}
		if strings.EqualFold(rawName, packageName) {
			lines[i] = matches[1] + matches[2] + matches[3] + matches[4] + matches[5] + version + matches[7]
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
	cmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", packageName, normalizeGoVersion(version)))
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update go.mod: %v\n%s", err, out)
	}
	return nil
}

// normalizeGoVersion prepends a "v" to bare semver strings (e.g. "3.50.0") so they
// are accepted by `go get pkg@version`. Values that already begin with "v", or that
// do not look like a bare semver (e.g. "latest", branch names, commit SHAs), are
// returned unchanged.
func normalizeGoVersion(version string) string {
	if version == "" || version[0] < '0' || version[0] > '9' || !strings.Contains(version, ".") {
		return version
	}
	return "v" + version
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

// updateCsprojFile upserts a PackageReference version in a .csproj file using proper XML parsing.
// It preserves all existing attributes (PrivateAssets, Condition, etc.) and avoids the
// ambiguity of inserting into the wrong ItemGroup that string-replace approaches suffer from.
func updateCsprojFile(path string, packageName string, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	trimmed := bytes.TrimPrefix(bytes.TrimLeft(data, " \t\r\n"), []byte{0xEF, 0xBB, 0xBF})
	hadXMLHeader := bytes.HasPrefix(bytes.TrimLeft(trimmed, " \t\r\n"), []byte("<?xml"))

	var root xmlNode
	if err := xml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse XML in %s: %w", path, err)
	}

	// Walk the tree looking for a matching PackageReference.
	if updated := upsertPackageReference(&root, packageName, version); !updated {
		// Not found — add a new PackageReference in a new ItemGroup.
		newRef := xmlNode{
			XMLName: xml.Name{Local: "PackageReference"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "Include"}, Value: packageName},
				{Name: xml.Name{Local: "Version"}, Value: version},
			},
		}
		root.Nodes = append(root.Nodes, xmlNode{
			XMLName: xml.Name{Local: "ItemGroup"},
			Nodes:   []xmlNode{newRef},
		})
	}

	output, err := xml.MarshalIndent(&root, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML for %s: %w", path, err)
	}
	if hadXMLHeader {
		output = append([]byte(xml.Header), output...)
	}

	if err := os.WriteFile(path, output, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// upsertPackageReference walks an xmlNode tree and updates the Version attribute of the
// first PackageReference whose Include attribute matches packageName (case-insensitive).
// Returns true if a match was found and updated.
func upsertPackageReference(node *xmlNode, packageName, version string) bool {
	for i := range node.Nodes {
		child := &node.Nodes[i]
		if child.XMLName.Local == "PackageReference" {
			for _, attr := range child.Attrs {
				if strings.EqualFold(attr.Name.Local, "Include") && strings.EqualFold(attr.Value, packageName) {
					// Found the entry — update the existing version representation if present.
					for k, a := range child.Attrs {
						if strings.EqualFold(a.Name.Local, "Version") {
							child.Attrs[k].Value = version
							return true
						}
					}

					for k := range child.Nodes {
						if strings.EqualFold(child.Nodes[k].XMLName.Local, "Version") {
							child.Nodes[k].Content = []byte(version)
							child.Nodes[k].Nodes = nil
							return true
						}
					}

					// No version metadata existed yet — append a Version attribute.
					child.Attrs = append(child.Attrs, xml.Attr{
						Name:  xml.Name{Local: "Version"},
						Value: version,
					})
					return true
				}
			}
		}
		if upsertPackageReference(child, packageName, version) {
			return true
		}
	}
	return false
}

// editDependencyYAML pins a provider version via the options.version of an explicit
// `pulumi:providers:<packageName>` resource; returns an error if no such resource exists.
func editDependencyYAML(workingDir string, packageName string, version string) error {
	pulumiYAMLPath := filepath.Join(workingDir, "Pulumi.yaml")
	data, err := os.ReadFile(pulumiYAMLPath)
	if err != nil {
		return fmt.Errorf("failed to read Pulumi.yaml: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse Pulumi.yaml: %w", err)
	}
	if doc.Kind == 0 {
		return fmt.Errorf("empty Pulumi.yaml")
	}
	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("root of Pulumi.yaml is not a mapping")
	}

	if err := yamlUpsertExplicitProviderVersion(root, packageName, version); err != nil {
		return fmt.Errorf("failed to update explicit provider in Pulumi.yaml: %w", err)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("failed to marshal Pulumi.yaml: %w", err)
	}
	if err := os.WriteFile(pulumiYAMLPath, out, 0644); err != nil {
		return fmt.Errorf("failed to write Pulumi.yaml: %w", err)
	}
	return nil
}

// yamlUpsertExplicitProviderVersion sets options.version on the `pulumi:providers:<packageName>`
// resource under `resources`; returns an error if that resource is not declared.
func yamlUpsertExplicitProviderVersion(root *yaml.Node, packageName, version string) error {
	resourcesNode := yamlMappingGet(root, "resources")
	if resourcesNode == nil || resourcesNode.Kind != yaml.MappingNode {
		return fmt.Errorf("no 'resources' mapping found; add an explicit provider resource of type pulumi:providers:%s", packageName)
	}

	wantType := "pulumi:providers:" + packageName
	for i := 0; i+1 < len(resourcesNode.Content); i += 2 {
		resValue := resourcesNode.Content[i+1]
		if resValue.Kind != yaml.MappingNode {
			continue
		}
		typeNode := yamlMappingGet(resValue, "type")
		if typeNode == nil || !strings.EqualFold(typeNode.Value, wantType) {
			continue
		}

		optionsNode := yamlMappingGet(resValue, "options")
		if optionsNode == nil {
			optionsNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			resValue.Content = append(resValue.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "options", Tag: "!!str"},
				optionsNode,
			)
		} else if optionsNode.Kind != yaml.MappingNode {
			return fmt.Errorf("'options' for explicit provider resource %q must be a mapping", wantType)
		}

		versionNode := yamlMappingGet(optionsNode, "version")
		if versionNode != nil {
			versionNode.Value = version
			versionNode.Tag = "!!str"
			versionNode.Style = 0
		} else {
			optionsNode.Content = append(optionsNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "version", Tag: "!!str"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: version, Tag: "!!str"},
			)
		}
		return nil
	}

	return fmt.Errorf("no explicit provider resource of type %q found in Pulumi.yaml; add one to pin the version", wantType)
}

// yamlMappingGet returns the value node for the given key in a YAML mapping node, or nil.
func yamlMappingGet(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
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
			if err := editDependencyYAML(workingDir, edit.PackageName, edit.Version); err != nil {
				return fmt.Errorf("failed to edit YAML dependency: %w", err)
			}
		default:
			return fmt.Errorf("unknown language detected: %s", language)
		}
	}

	return nil
}
