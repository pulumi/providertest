package autotest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (a *AutoTest) Convert(language string) (*AutoTest, string) {
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

	return a.WithSource(targetDir), string(out)
}
