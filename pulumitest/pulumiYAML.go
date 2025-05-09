package pulumitest

import (
	"os"
	"path/filepath"
	"strings"
)

// WritePulumiYaml writes the contents of the program string to the Pulumi.yaml file in the current testing directory.
// YAML does not allow tabs, so this function will error if the program contains tabs.
func (pt *PulumiTest) WritePulumiYaml(t PT, program string) {
	t.Helper()

	// find the line of the program that contains tabs
	lines := strings.Split(program, "\n")
	for i, line := range lines {
		if strings.Contains(line, "\t") {
			ptFatalF(t, "program contains tabs on line %d: %s\nTry replacing it with:\n%s", i+1, line, strings.ReplaceAll(program, "\t", "    "))
		}
	}

	pulumiYamlPath := filepath.Join(pt.CurrentStack().Workspace().WorkDir(), "Pulumi.yaml")
	err := os.WriteFile(pulumiYamlPath, []byte(program), 0o600)
	if err != nil {
		ptFatalF(t, "failed to replace program %s", err)
	}
}

// ReadPulumiYaml reads the contents of the Pulumi.yaml file in the current testing directory.
func (pt *PulumiTest) ReadPulumiYaml(t PT) string {
	t.Helper()

	pulumiYamlPath := filepath.Join(pt.CurrentStack().Workspace().WorkDir(), "Pulumi.yaml")
	program, err := os.ReadFile(pulumiYamlPath)
	if err != nil {
		ptFatalF(t, "failed to read program %s", err)
	}
	return string(program)
}
