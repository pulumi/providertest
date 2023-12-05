package providerfactory

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func DownloadPluginBinary(t *testing.T, name, version string) string {
	t.Helper()

	cmd := exec.Command("pulumi", "plugin", "install", "resource", name, version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to install plugin: %s\n%s", err, out)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(userHomeDir, ".pulumi", "plugins", fmt.Sprintf("resource-%s-v%s", name, version))
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("expected plugin binary to exist at %s", binaryPath)
	}
	return binaryPath
}
