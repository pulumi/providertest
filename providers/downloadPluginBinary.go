package providers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func DownloadPluginBinaryFactory(name, version string) ProviderFactory {
	factory := func(ctx context.Context) (Port, error) {
		binaryPath, err := DownloadPluginBinary(name, version)
		if err != nil {
			return 0, err
		}
		return startLocalBinary(ctx, binaryPath, name)
	}
	return factory
}

func DownloadPluginBinary(name, version string) (string, error) {
	cmd := exec.Command("pulumi", "plugin", "install", "resource", name, version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to install plugin: %s\n%s", err, out)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %v", err)
	}
	binaryPath := filepath.Join(userHomeDir, ".pulumi", "plugins", fmt.Sprintf("resource-%s-v%s", name, version))
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("expected plugin binary to exist at %s", binaryPath)
	}
	return binaryPath, nil
}
