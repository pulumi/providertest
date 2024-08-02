package pulumitest

import (
	"os"
	"path/filepath"
	"strings"
)

func (a *PulumiTest) ReplaceProgram(program string) {
	a.t.Helper()

	// YAML doesn't allow tabs but go uses tabs which makes for a miserable experience with inline yaml programs
	program = strings.ReplaceAll(program, "\t", "    ")

	pulumiYamlPath := filepath.Join(a.CurrentStack().Workspace().WorkDir(), "Pulumi.yaml")
	err := os.WriteFile(pulumiYamlPath, []byte(program), 0o600)
	if err != nil {
		a.fatalf("failed to replace program %s", err)
	}
}
