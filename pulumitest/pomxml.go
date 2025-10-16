package pulumitest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// findPomFile locates a pom.xml file in the given directory.
func findPomFile(dir string) (string, error) {
	// Check if pom.xml exists in the current directory
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		return pomPath, nil
	}

	// Return error if not found
	return "", fmt.Errorf("pom.xml not found in directory %s", dir)
}

// readPomXML reads and parses a pom.xml file into an xmlNode structure.
func readPomXML(pomPath string) (*xmlNode, error) {
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pom.xml: %w", err)
	}

	buf := bytes.NewBuffer(content)
	dec := xml.NewDecoder(buf)

	var root xmlNode
	err = dec.Decode(&root)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	return &root, nil
}

// writePomXML writes an xmlNode structure back to a pom.xml file.
func writePomXML(pomPath string, root *xmlNode) error {
	content, err := xml.MarshalIndent(root, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal pom.xml: %w", err)
	}

	// Add XML declaration
	fullContent := append([]byte(xml.Header), content...)

	err = os.WriteFile(pomPath, fullContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write pom.xml: %w", err)
	}

	return nil
}

// createDependencyNodes creates XML nodes for a Maven dependency.
func createDependencyNodes(groupID, artifactID, version, systemPath string) []xmlNode {
	nodes := []xmlNode{
		{XMLName: xml.Name{Local: "groupId"}, Content: []byte(groupID)},
		{XMLName: xml.Name{Local: "artifactId"}, Content: []byte(artifactID)},
		{XMLName: xml.Name{Local: "version"}, Content: []byte(version)},
	}

	if systemPath != "" {
		nodes = append(nodes,
			xmlNode{XMLName: xml.Name{Local: "scope"}, Content: []byte("system")},
			xmlNode{XMLName: xml.Name{Local: "systemPath"}, Content: []byte(systemPath)},
		)
	}

	return nodes
}

func setChildText(nodes *[]xmlNode, tagName, value string) {
	child := findChild(*nodes, tagName)
	if child != nil {
		child.Content = []byte(value)
		child.Nodes = nil
		return
	}

	*nodes = append(*nodes, xmlNode{
		XMLName: xml.Name{Local: tagName},
		Content: []byte(value),
	})
}

func removeChild(nodes *[]xmlNode, tagName string) {
	filtered := (*nodes)[:0]
	for _, node := range *nodes {
		if node.XMLName.Local != tagName {
			filtered = append(filtered, node)
		}
	}
	*nodes = filtered
}

// addOrUpdateDependency adds or updates a Maven dependency in pom.xml.
// If the dependency already exists, it will be updated; otherwise, a new one will be added.
// If systemPath is provided, it will be validated to ensure the file exists.
func addOrUpdateDependency(pomPath string, groupID, artifactID, version, systemPath string) error {
	if groupID == "" || artifactID == "" || version == "" {
		return fmt.Errorf("groupID, artifactID, and version cannot be empty")
	}

	// Validate systemPath if provided
	if systemPath != "" {
		if _, err := os.Stat(systemPath); err != nil {
			return fmt.Errorf("systemPath %s does not exist: %w", systemPath, err)
		}
	}

	root, err := readPomXML(pomPath)
	if err != nil {
		return err
	}

	// Find or create dependencies section
	dependenciesNode := findChild(root.Nodes, "dependencies")
	if dependenciesNode == nil {
		// Create new dependencies node
		newDependencies := xmlNode{
			XMLName: xml.Name{Local: "dependencies"},
		}
		root.Nodes = append(root.Nodes, newDependencies)
		dependenciesNode = &root.Nodes[len(root.Nodes)-1]
	}

	// Look for existing dependency
	dependencies := findAllChildren(dependenciesNode.Nodes, "dependency")
	for _, dep := range dependencies {
		groupNode := findChild(dep.Nodes, "groupId")
		artifactNode := findChild(dep.Nodes, "artifactId")

		if groupNode != nil && artifactNode != nil {
			// Extract text content from nodes
			existingGroup := getNodeTextContent(groupNode)
			existingArtifact := getNodeTextContent(artifactNode)

			// If we found a match, update it
			if existingGroup == groupID && existingArtifact == artifactID {
				setChildText(&dep.Nodes, "groupId", groupID)
				setChildText(&dep.Nodes, "artifactId", artifactID)
				setChildText(&dep.Nodes, "version", version)
				if systemPath != "" {
					setChildText(&dep.Nodes, "scope", "system")
					setChildText(&dep.Nodes, "systemPath", systemPath)
				} else {
					removeChild(&dep.Nodes, "scope")
					removeChild(&dep.Nodes, "systemPath")
				}
				dep.Content = nil // Clear content to avoid duplication during marshaling
				return writePomXML(pomPath, root)
			}
		}
	}

	// Create new dependency
	newDependency := xmlNode{
		XMLName: xml.Name{Local: "dependency"},
		Nodes:   createDependencyNodes(groupID, artifactID, version, systemPath),
	}

	dependenciesNode.Nodes = append(dependenciesNode.Nodes, newDependency)
	return writePomXML(pomPath, root)
}

func updateCompilerPluginJavaVersion(root *xmlNode, version string) {
	buildNode := findChild(root.Nodes, "build")
	if buildNode == nil {
		return
	}
	pluginsNode := findChild(buildNode.Nodes, "plugins")
	if pluginsNode == nil {
		return
	}

	for _, plugin := range findAllChildren(pluginsNode.Nodes, "plugin") {
		artifactNode := findChild(plugin.Nodes, "artifactId")
		if artifactNode == nil || strings.TrimSpace(getNodeTextContent(artifactNode)) != "maven-compiler-plugin" {
			continue
		}

		configurationNode := findChild(plugin.Nodes, "configuration")
		if configurationNode == nil {
			plugin.Nodes = append(plugin.Nodes, xmlNode{XMLName: xml.Name{Local: "configuration"}})
			configurationNode = &plugin.Nodes[len(plugin.Nodes)-1]
		}

		setChildText(&configurationNode.Nodes, "source", version)
		setChildText(&configurationNode.Nodes, "target", version)
	}
}

// setJavaVersion sets the Java compiler source and target version in pom.xml.
// It updates maven.compiler.source and maven.compiler.target properties.
// If the properties section doesn't exist, it will be created.
// Existing properties are updated in place (preserving their original order);
// new properties are appended at the end in a deterministic order.
func setJavaVersion(pomPath string, version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	root, err := readPomXML(pomPath)
	if err != nil {
		return err
	}

	// Find or create properties section
	propertiesNode := findChild(root.Nodes, "properties")
	if propertiesNode == nil {
		newProperties := xmlNode{
			XMLName: xml.Name{Local: "properties"},
		}
		root.Nodes = append(root.Nodes, newProperties)
		propertiesNode = &root.Nodes[len(root.Nodes)-1]
	}

	// Properties to update/create
	propsToSet := map[string]string{
		"maven.compiler.source": version,
		"maven.compiler.target": version,
	}

	// Update existing properties in place, preserving user-defined order.
	for i := range propertiesNode.Nodes {
		if newValue, ok := propsToSet[propertiesNode.Nodes[i].XMLName.Local]; ok {
			propertiesNode.Nodes[i].Content = []byte(newValue)
			delete(propsToSet, propertiesNode.Nodes[i].XMLName.Local)
		}
	}

	// Append any missing properties in a deterministic order.
	missingNames := make([]string, 0, len(propsToSet))
	for name := range propsToSet {
		missingNames = append(missingNames, name)
	}
	sort.Strings(missingNames)
	for _, propName := range missingNames {
		propertiesNode.Nodes = append(propertiesNode.Nodes, xmlNode{
			XMLName: xml.Name{Local: propName},
			Content: []byte(propsToSet[propName]),
		})
	}

	updateCompilerPluginJavaVersion(root, version)

	return writePomXML(pomPath, root)
}
