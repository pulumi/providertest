package autotest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Convert(t *testing.T, ctx context.Context, source, language string) string {
	t.Helper()

	tempDir := t.TempDir()
	base := filepath.Base(source)
	targetDir := filepath.Join(tempDir, fmt.Sprintf("%s-%s", base, language))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", targetDir)
	cmd.Dir = source
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to convert directory: %s\n%s", err, out)
	}

	return targetDir
}
