package pulumitest

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findCsprojFile finds a .csproj file in the given directory
func findCsprojFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".csproj") {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no .csproj file found in directory %s", dir)
}

// addProjectReferences adds ProjectReference elements to a .csproj file
func addProjectReferences(csprojPath string, references map[string]string) error {
	// Read and parse the .csproj file
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return fmt.Errorf("failed to read .csproj file: %w", err)
	}

	var root xmlNode
	if err := xml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse .csproj XML: %w", err)
	}

	// Build ProjectReference nodes
	var refNodes []xmlNode
	for _, refPath := range references {
		absPath, err := filepath.Abs(refPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", refPath, err)
		}

		// Check if the path exists
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("reference path does not exist: %s: %w", absPath, err)
		}

		// If it's a directory, look for a .csproj file in it
		if info.IsDir() {
			absPath, err = findCsprojFile(absPath)
			if err != nil {
				return fmt.Errorf("failed to find .csproj in directory: %w", err)
			}
		}

		refNodes = append(refNodes, xmlNode{
			XMLName: xml.Name{Local: "ProjectReference"},
			Attrs:   []xml.Attr{{Name: xml.Name{Local: "Include"}, Value: absPath}},
		})
	}

	// Create new ItemGroup node with ProjectReferences
	itemGroup := xmlNode{
		XMLName: xml.Name{Local: "ItemGroup"},
		Nodes:   refNodes,
	}

	// Add the new ItemGroup to the root
	root.Nodes = append(root.Nodes, itemGroup)

	// Marshal back to XML
	output, err := xml.MarshalIndent(&root, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal .csproj XML: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(csprojPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write .csproj file: %w", err)
	}

	return nil
}
