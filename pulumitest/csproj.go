package pulumitest

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CsprojProject represents a simplified .csproj XML structure
type CsprojProject struct {
	XMLName xml.Name `xml:"Project"`
	Sdk     string   `xml:"Sdk,attr,omitempty"`
	Groups  []CsprojItemGroup
}

// CsprojItemGroup represents an ItemGroup in .csproj
type CsprojItemGroup struct {
	XMLName            xml.Name              `xml:"ItemGroup"`
	PackageReferences  []CsprojPackageRef    `xml:"PackageReference"`
	ProjectReferences  []CsprojProjectRef    `xml:"ProjectReference"`
	PropertyGroups     []CsprojPropertyGroup `xml:"PropertyGroup"`
}

// CsprojPackageRef represents a PackageReference element
type CsprojPackageRef struct {
	XMLName xml.Name `xml:"PackageReference"`
	Include string   `xml:"Include,attr"`
	Version string   `xml:"Version,attr,omitempty"`
}

// CsprojProjectRef represents a ProjectReference element
type CsprojProjectRef struct {
	XMLName xml.Name `xml:"ProjectReference"`
	Include string   `xml:"Include,attr"`
}

// CsprojPropertyGroup represents a PropertyGroup element
type CsprojPropertyGroup struct {
	XMLName xml.Name `xml:"PropertyGroup"`
}

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
	// Read the existing .csproj file
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return fmt.Errorf("failed to read .csproj file: %w", err)
	}

	// Simple XML manipulation approach:
	// We'll parse the file as text and insert new ProjectReference elements
	// before the closing </Project> tag
	content := string(data)

	// Build the ProjectReference XML elements
	var projectRefs strings.Builder
	projectRefs.WriteString("\n\n  <ItemGroup>\n")

	for _, refPath := range references {
		absPath, err := filepath.Abs(refPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", refPath, err)
		}

		// Check if the path exists
		_, err = os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("reference path does not exist: %s: %w", absPath, err)
		}

		// If it's a directory, look for a .csproj file in it
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("failed to stat path %s: %w", absPath, err)
		}

		if info.IsDir() {
			absPath, err = findCsprojFile(absPath)
			if err != nil {
				return fmt.Errorf("failed to find .csproj in directory: %w", err)
			}
		}

		projectRefs.WriteString(fmt.Sprintf("    <ProjectReference Include=\"%s\" />\n", absPath))
	}

	projectRefs.WriteString("  </ItemGroup>\n")

	// Find the closing </Project> tag and insert before it
	closeTagIndex := strings.LastIndex(content, "</Project>")
	if closeTagIndex == -1 {
		return fmt.Errorf("malformed .csproj file: no closing </Project> tag found")
	}

	// Insert the new ItemGroup before the closing tag
	newContent := content[:closeTagIndex] + projectRefs.String() + content[closeTagIndex:]

	// Write back to file
	err = os.WriteFile(csprojPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .csproj file: %w", err)
	}

	return nil
}

// setTargetFramework modifies the TargetFramework in a .csproj file
func setTargetFramework(csprojPath string, framework string) error {
	// Read the existing .csproj file
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return fmt.Errorf("failed to read .csproj file: %w", err)
	}

	content := string(data)

	// Find and replace the TargetFramework element
	// Look for <TargetFramework>...</TargetFramework>
	startTag := "<TargetFramework>"
	endTag := "</TargetFramework>"

	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return fmt.Errorf("no <TargetFramework> element found in .csproj file")
	}

	endIndex := strings.Index(content[startIndex:], endTag)
	if endIndex == -1 {
		return fmt.Errorf("malformed <TargetFramework> element in .csproj file")
	}

	// Calculate the actual end index in the full content
	endIndex = startIndex + endIndex

	// Replace the framework value
	newContent := content[:startIndex+len(startTag)] + framework + content[endIndex:]

	// Write back to file
	err = os.WriteFile(csprojPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .csproj file: %w", err)
	}

	return nil
}
