package pulumitest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPomFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a pom.xml file
	pomPath := filepath.Join(tempDir, "pom.xml")
	err := os.WriteFile(pomPath, []byte("<project></project>"), 0644)
	require.NoError(t, err)

	// Find the pom.xml
	found, err := findPomFile(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, pomPath, found)
}

func TestFindPomFileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	// Try to find a pom.xml that doesn't exist
	_, err := findPomFile(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pom.xml not found")
}

func TestAddOrUpdateDependency(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a basic pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
        </dependency>
    </dependencies>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Add a new dependency
	err = addOrUpdateDependency(pomPath, "com.pulumi", "pulumi", "0.17.0", "")
	assert.NoError(t, err)

	// Read the file and verify
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify the new dependency is present
	assert.Contains(t, contentStr, "com.pulumi")
	assert.Contains(t, contentStr, "pulumi")
	assert.Contains(t, contentStr, "0.17.0")

	// Verify the old dependency is still there
	assert.Contains(t, contentStr, "junit")
}

func TestAddOrUpdateDependencyWithSystemPath(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a basic pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
    </dependencies>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Create a temporary JAR file for systemPath
	jarPath := filepath.Join(tempDir, "pulumi.jar")
	err = os.WriteFile(jarPath, []byte("fake jar content"), 0644)
	require.NoError(t, err)

	// Add a dependency with system path
	err = addOrUpdateDependency(pomPath, "com.pulumi", "pulumi", "0.0.0-dev", jarPath)
	assert.NoError(t, err)

	// Read the file and verify
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify system scope and path are present
	assert.Contains(t, contentStr, "system")
	assert.Contains(t, contentStr, jarPath)
}


func TestUpdateExistingDependency(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a pom.xml with an existing dependency
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>com.pulumi</groupId>
            <artifactId>pulumi</artifactId>
            <version>0.16.0</version>
        </dependency>
    </dependencies>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Create a temporary JAR file for systemPath
	jarPath := filepath.Join(tempDir, "local.jar")
	err = os.WriteFile(jarPath, []byte("fake jar content"), 0644)
	require.NoError(t, err)

	// Update existing dependency with new version and system path
	err = addOrUpdateDependency(pomPath, "com.pulumi", "pulumi", "0.17.0", jarPath)
	assert.NoError(t, err)

	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify updated version
	assert.Contains(t, contentStr, "0.17.0")
	assert.NotContains(t, contentStr, "0.16.0")
	assert.Contains(t, contentStr, "system")
	assert.Contains(t, contentStr, "local.jar")
}

func TestSetJavaVersion(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a pom.xml without properties
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <modelVersion>4.0.0</modelVersion>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Set Java version
	err = setJavaVersion(pomPath, "17")
	assert.NoError(t, err)

	// Read the file and verify
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify properties were added
	assert.Contains(t, contentStr, "maven.compiler.source")
	assert.Contains(t, contentStr, "maven.compiler.target")
	assert.Contains(t, contentStr, "<maven.compiler.source>17</maven.compiler.source>")
	assert.Contains(t, contentStr, "<maven.compiler.target>17</maven.compiler.target>")
}

func TestSetJavaVersionUpdateExisting(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a pom.xml with existing properties
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Update Java version
	err = setJavaVersion(pomPath, "17")
	assert.NoError(t, err)

	// Read the file and verify
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify properties were updated
	assert.Contains(t, contentStr, "<maven.compiler.source>17</maven.compiler.source>")
	assert.Contains(t, contentStr, "<maven.compiler.target>17</maven.compiler.target>")
	// Verify other properties are preserved
	assert.Contains(t, contentStr, "project.build.sourceEncoding")
}


func TestReadWritePomXML(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a simple pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>example</artifactId>
    <version>1.0.0</version>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Read the file
	root, err := readPomXML(pomPath)
	assert.NoError(t, err)
	assert.NotNil(t, root)

	// Verify we can access child nodes
	assert.Equal(t, "project", root.XMLName.Local)

	// Modify and write back
	err = writePomXML(pomPath, root)
	assert.NoError(t, err)

	// Verify file was written
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	assert.True(t, len(content) > 0)
	assert.Contains(t, string(content), "<?xml version")
}

func TestGetNodeTextContent(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a pom.xml with text content
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Read and find groupId node
	root, err := readPomXML(pomPath)
	assert.NoError(t, err)

	groupNode := findChild(root.Nodes, "groupId")
	assert.NotNil(t, groupNode)

	// Extract text content
	text := getNodeTextContent(groupNode)
	assert.Equal(t, "com.example", text)
}

func TestPreserveXMLNamespaces(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Create a pom.xml with XML namespace (like real Maven projects)
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test</artifactId>
    <version>1.0.0</version>
</project>`

	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Modify the pom.xml by setting Java version
	err = setJavaVersion(pomPath, "17")
	assert.NoError(t, err)

	// Read the file and verify main namespace is preserved
	content, err := os.ReadFile(pomPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Verify the main XML namespace is preserved
	assert.Contains(t, contentStr, `xmlns="http://maven.apache.org/POM/4.0.0"`)

	// Verify the modification worked
	assert.Contains(t, contentStr, "<maven.compiler.source>17</maven.compiler.source>")
	assert.Contains(t, contentStr, "<maven.compiler.target>17</maven.compiler.target>")

	// Verify existing elements are still present (content, not exact format)
	assert.Contains(t, contentStr, "<modelVersion")
	assert.Contains(t, contentStr, "4.0.0</modelVersion>")
	assert.Contains(t, contentStr, "<groupId")
	assert.Contains(t, contentStr, "com.example</groupId>")
}

func TestAddDependencyWithEmptyInputs(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies></dependencies>
</project>`
	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name       string
		groupID    string
		artifactID string
		version    string
		wantErr    string
	}{
		{"empty groupID", "", "artifact", "1.0", "groupID"},
		{"empty artifactID", "group", "", "1.0", "artifactID"},
		{"empty version", "group", "artifact", "", "version"},
		{"all empty", "", "", "", "groupID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addOrUpdateDependency(pomPath, tt.groupID, tt.artifactID, tt.version, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSetJavaVersionEmpty(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")
	err := os.WriteFile(pomPath, []byte("<project></project>"), 0644)
	require.NoError(t, err)

	err = setJavaVersion(pomPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestReadPomXMLInvalidXML(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")

	// Write invalid XML
	err := os.WriteFile(pomPath, []byte("<project><unclosed>"), 0644)
	require.NoError(t, err)

	_, err = readPomXML(pomPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse pom.xml")
}

func TestReadPomXMLNonExistent(t *testing.T) {
	_, err := readPomXML("/nonexistent/pom.xml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read pom.xml")
}

func TestAddDependencyWithInvalidSystemPath(t *testing.T) {
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies></dependencies>
</project>`
	err := os.WriteFile(pomPath, []byte(pomContent), 0644)
	require.NoError(t, err)

	// Try to add dependency with non-existent systemPath
	err = addOrUpdateDependency(pomPath, "com.example", "test", "1.0", "/nonexistent/path.jar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "systemPath")
	assert.Contains(t, err.Error(), "does not exist")
}
