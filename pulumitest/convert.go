package pulumitest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ConvertResult struct {
	// AutoTest instance for the converted program.
	AutoTest *PulumiTest
	// Combined output of the `pulumi convert` command.
	Output string
}

// Convert a program to a given language.
// It returns a new AutoTest instance for the converted program which will be outputted into a temporary directory.
func (a *PulumiTest) Convert(language string) ConvertResult {
	a.t.Helper()

	tempDir := a.t.TempDir()
	base := filepath.Base(a.source)
	targetDir := filepath.Join(tempDir, fmt.Sprintf("%s-%s", base, language))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		a.t.Fatal(err)
	}

	a.t.Logf("converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", targetDir)
	cmd.Dir = a.source
	out, err := cmd.CombinedOutput()
	if err != nil {
		a.t.Fatalf("failed to convert directory: %s\n%s", err, out)
	}

	return ConvertResult{
		AutoTest: &PulumiTest{
			t:       a.t,
			ctx:     a.ctx,
			source:  targetDir,
			options: a.options,
		},
		Output: string(out),
	}
}
