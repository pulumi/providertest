package pulumitest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest/optnewstack"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremove"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// NewStack creates a new stack, ensure it's cleaned up after the test is done.
// If no stack name is provided, a random one will be generated.
func (pt *PulumiTest) NewStack(stackName string, opts ...optnewstack.NewStackOpt) *auto.Stack {
	pt.t.Helper()

	stackOptions := optnewstack.Defaults()
	for _, opt := range opts {
		opt.Apply(&stackOptions)
	}

	if stackName == "" {
		stackName = randomStackName(pt.source)
	}

	options := pt.options

	// Set default stack opts. These can be overridden by the caller.
	env := map[string]string{}

	if options.ConfigPassphrase != "" {
		env["PULUMI_CONFIG_PASSPHRASE"] = options.ConfigPassphrase
	}

	if !options.UseAmbientBackend {
		backendFolder := pt.t.TempDir()
		env["PULUMI_BACKEND_URL"] = "file://" + backendFolder
	}

	if !options.DisableGrpcLog {
		grpcLogDir := pt.t.TempDir()
		env["PULUMI_DEBUG_GRPC"] = filepath.Join(grpcLogDir, "grpc.json")
	}

	providerFactories := options.ProviderFactories()
	if len(providerFactories) > 0 {
		pt.t.Log("starting providers")
		providerContext, cancelProviders := context.WithCancel(pt.ctx)
		providerPorts, err := providers.StartProviders(providerContext, providerFactories, pt)
		if err != nil {
			cancelProviders()
			pt.t.Fatalf("failed to start providers: %v", err)
		} else {
			pt.t.Cleanup(func() {
				cancelProviders()
			})
		}
		env["PULUMI_DEBUG_PROVIDERS"] = providers.GetDebugProvidersEnv(providerPorts)
	}

	// Apply custom env last to allow overriding any of the above.
	for k, v := range options.CustomEnv {
		env[k] = v
	}

	stackOpts := []auto.LocalWorkspaceOption{
		auto.EnvVars(env),
	}
	stackOpts = append(stackOpts, options.ExtraWorkspaceOptions...)
	stackOpts = append(stackOpts, stackOptions.Opts...)

	pt.PT().Logf("creating stack %s", stackName)
	stack, err := auto.NewStackLocalSource(pt.ctx, stackName, pt.source, stackOpts...)

	providerPluginPaths := options.ProviderPluginPaths()
	if len(providerPluginPaths) > 0 {
		projectSettings, err := stack.Workspace().ProjectSettings(pt.ctx)
		if err != nil {
			pt.t.Fatalf("failed to get project settings: %s", err)
		}
		var plugins workspace.Plugins
		if projectSettings.Plugins != nil {
			plugins = *projectSettings.Plugins
		}
		providerPlugins := plugins.Providers
		// Sort the provider plugin paths to ensure a consistent order.
		providerPluginNames := make([]string, 0, len(providerPluginPaths))
		for name := range providerPluginPaths {
			providerPluginNames = append(providerPluginNames, string(name))
		}
		sort.Strings(providerPluginNames)
		for _, name := range providerPluginNames {
			relPath := providerPluginPaths[providers.ProviderName(name)]
			absPath, err := filepath.Abs(relPath)
			if err != nil {
				pt.t.Fatalf("failed to get absolute path for %s: %s", relPath, err)
			}

			found := false
			for _, provider := range providerPlugins {
				if provider.Name == name {
					provider.Path = absPath
					found = true
					break
				}
			}
			if !found {
				providerPlugins = append(providerPlugins, workspace.PluginOptions{
					Name: name,
					Path: absPath,
				})
			}
		}
		plugins.Providers = providerPlugins
		projectSettings.Plugins = &plugins
		err = stack.Workspace().SaveProjectSettings(pt.ctx, projectSettings)
		if err != nil {
			pt.t.Fatalf("failed to save project settings: %s", err)
		}
	}

	if options.YarnLinks != nil && len(options.YarnLinks) > 0 {
		for _, pkg := range options.YarnLinks {
			cmd := exec.Command("yarn", "link", pkg)
			cmd.Dir = pt.source
			pt.t.Logf("linking yarn package: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				pt.t.Fatalf("failed to link yarn package %s: %s\n%s", pkg, err, out)
			}
		}
	}

	if options.GoModReplacements != nil && len(options.GoModReplacements) > 0 {
		orderedReplacements := make([]string, 0, len(options.GoModReplacements))
		for old := range options.GoModReplacements {
			orderedReplacements = append(orderedReplacements, old)
		}
		sort.Strings(orderedReplacements)
		for _, old := range orderedReplacements {
			relPath := options.GoModReplacements[old]
			absPath, err := filepath.Abs(relPath)
			if err != nil {
				pt.t.Fatalf("failed to get absolute path for %s: %s", relPath, err)
			}
			replacement := fmt.Sprintf("%s=%s", old, absPath)
			cmd := exec.Command("go", "mod", "edit", "-replace", replacement)
			cmd.Dir = pt.source
			pt.t.Logf("adding go.mod replacement: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				pt.t.Fatalf("failed to add go.mod replacement %s: %s\n%s", replacement, err, out)
			}
		}
	}

	if err != nil {
		pt.t.Fatalf("failed to create stack: %s", err)
		return nil
	}
	if !stackOptions.SkipDestroy {
		pt.t.Cleanup(func() {
			pt.t.Helper()
			pt.t.Log("cleaning up stack")
			_, err := stack.Destroy(pt.ctx)
			if err != nil {
				pt.t.Errorf("failed to destroy stack: %s", err)
			}
			err = stack.Workspace().RemoveStack(pt.ctx, stackName, optremove.Force())
			if err != nil {
				pt.t.Errorf("failed to remove stack: %s", err)
			}
		})
	}
	pt.currentStack = &stack
	return &stack
}

func randomStackName(dir string) string {
	// Fetch the host and test dir names, cleaned so to contain just [a-zA-Z0-9-_] chars.
	hostname, err := os.Hostname()
	contract.AssertNoErrorf(err, "failure to fetch hostname for stack prefix")
	var host string
	for _, c := range hostname {
		if len(host) >= 10 {
			break
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			host += string(c)
		}
	}

	var test string
	for _, c := range filepath.Base(dir) {
		if len(test) >= 10 {
			break
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			test += string(c)
		}
	}

	b := make([]byte, 4)
	_, err = rand.Read(b)
	contract.AssertNoErrorf(err, "failure to generate random stack suffix")

	return strings.ToLower("p-it-" + host + "-" + test + "-" + hex.EncodeToString(b))

}
