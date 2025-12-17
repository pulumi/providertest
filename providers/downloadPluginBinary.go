package providers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func DownloadPluginBinaryFactory(name, version string) ProviderFactory {
	factory := func(ctx context.Context, pt PulumiTest) (Port, error) {
		binaryPath, err := DownloadPluginBinary(name, version)
		if err != nil {
			return 0, err
		}
		return startLocalBinary(ctx, binaryPath, name, pt.Source())
	}
	return factory
}

func DownloadPluginBinary(name, version string) (string, error) {
	cmd := exec.Command("pulumi", "plugin", "install", "resource", name, version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to install plugin: %s\n%s", err, out)
	}

	pluginDir, err := workspace.GetPluginDir()
	if err != nil {
		return "", fmt.Errorf("failed to get plugin dir: %v", err)
	}

	binaryPath := filepath.Join(pluginDir, fmt.Sprintf("resource-%s-v%s", name, version))
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("expected plugin binary to exist at %s", binaryPath)
	}
	return binaryPath, nil
}
