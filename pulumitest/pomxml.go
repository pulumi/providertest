package pulumitest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
				// Replace dependency nodes with updated ones
				dep.Nodes = createDependencyNodes(groupID, artifactID, version, systemPath)
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

// setJavaVersion sets the Java compiler source and target version in pom.xml.
// It updates maven.compiler.source and maven.compiler.target properties.
// If the properties section doesn't exist, it will be created.
// Existing properties will be updated, new properties will be added.
// Note: Properties may be reordered alphabetically for consistency.
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

	// Update existing properties
	for i := range propertiesNode.Nodes {
		if propsToSet[propertiesNode.Nodes[i].XMLName.Local] != "" {
			propertiesNode.Nodes[i].Content = []byte(propsToSet[propertiesNode.Nodes[i].XMLName.Local])
			delete(propsToSet, propertiesNode.Nodes[i].XMLName.Local)
		}
	}

	// Add missing properties
	for propName, propValue := range propsToSet {
		propertiesNode.Nodes = append(propertiesNode.Nodes, xmlNode{
			XMLName: xml.Name{Local: propName},
			Content: []byte(propValue),
		})
	}

	// Sort properties for consistency
	sort.Slice(propertiesNode.Nodes, func(i, j int) bool {
		return propertiesNode.Nodes[i].XMLName.Local < propertiesNode.Nodes[j].XMLName.Local
	})

	return writePomXML(pomPath, root)
}
