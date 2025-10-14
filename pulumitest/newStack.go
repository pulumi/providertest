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
func (pt *PulumiTest) NewStack(t PT, stackName string, opts ...optnewstack.NewStackOpt) *auto.Stack {
	t.Helper()

	stackOptions := optnewstack.Defaults()
	for _, opt := range opts {
		opt.Apply(&stackOptions)
	}

	if stackName == "" {
		stackName = randomStackName(pt.workingDir)
	}

	options := pt.options

	// Set default stack opts. These can be overridden by the caller.
	env := map[string]string{}

	if options.ConfigPassphrase != "" {
		env["PULUMI_CONFIG_PASSPHRASE"] = options.ConfigPassphrase
	}

	if !options.UseAmbientBackend {
		backendFolder := tempDirWithoutCleanupOnFailedTest(t, "backendDir", options.TempDir)
		t.Log("PULUMI_BACKEND_URL=" + "file://" + backendFolder)
		env["PULUMI_BACKEND_URL"] = "file://" + backendFolder
	}

	if !options.DisableGrpcLog {
		grpcLogDir := tempDirWithoutCleanupOnFailedTest(t, "grpcLogDir", options.TempDir)
		t.Log("PULUMI_DEBUG_GRPC=" + filepath.Join(grpcLogDir, "grpc.json"))
		env["PULUMI_DEBUG_GRPC"] = filepath.Join(grpcLogDir, "grpc.json")
	}

	providerFactories := options.ProviderFactories()
	if len(providerFactories) > 0 {
		t.Log("starting providers")
		providerContext, cancelProviders := context.WithCancel(pt.ctx)
		providerPorts, err := providers.StartProviders(providerContext, providerFactories, pt)
		if err != nil {
			cancelProviders()
			ptFatalF(t, "failed to start providers: %v", err)
		} else {
			t.Cleanup(func() {
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

	ptLogF(t, "creating stack %s", stackName)
	stack, err := auto.NewStackLocalSource(pt.ctx, stackName, pt.workingDir, stackOpts...)

	providerPluginPaths := options.ProviderPluginPaths()
	if len(providerPluginPaths) > 0 {
		projectSettings, err := stack.Workspace().ProjectSettings(pt.ctx)
		if err != nil {
			ptFatalF(t, "failed to get project settings: %s", err)
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
				ptFatalF(t, "failed to get absolute path for %s: %s", relPath, err)
			}
			_, err = os.Stat(absPath)
			if err != nil {
				ptFatalF(t, "failed to find binary for provider %q: %s", name, err)
				return nil
			}

			found := false
			for idx := range providerPlugins {
				if providerPlugins[idx].Name == name {
					providerPlugins[idx].Path = absPath
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
			ptFatalF(t, "failed to save project settings: %s", err)
		}
	}

	if options.YarnLinks != nil && len(options.YarnLinks) > 0 {
		for _, pkg := range options.YarnLinks {
			cmd := exec.Command("yarn", "link", pkg)
			cmd.Dir = pt.workingDir
			ptLogF(t, "linking yarn package: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				ptFatalF(t, "failed to link yarn package %s: %s\n%s", pkg, err, out)
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
				ptFatalF(t, "failed to get absolute path for %s: %s", relPath, err)
			}
			replacement := fmt.Sprintf("%s=%s", old, absPath)
			cmd := exec.Command("go", "mod", "edit", "-replace", replacement)
			cmd.Dir = pt.workingDir
			ptLogF(t, "adding go.mod replacement: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				ptFatalF(t, "failed to add go.mod replacement %s: %s\n%s", replacement, err, out)
			}
		}
	}

	// Handle .NET-specific configurations
	if len(options.DotNetReferences) > 0 || options.DotNetTargetFramework != "" || options.DotNetBuildConfig != "" {
		// Find the .csproj file in the working directory
		csprojPath, err := findCsprojFile(pt.workingDir)
		if err != nil {
			ptFatalF(t, "failed to find .csproj file: %s", err)
		}
		ptLogF(t, "found .csproj file: %s", csprojPath)

		// Add project references
		if len(options.DotNetReferences) > 0 {
			err = addProjectReferences(csprojPath, options.DotNetReferences)
			if err != nil {
				ptFatalF(t, "failed to add .NET project references: %s", err)
			}

			// Log the references that were added
			for name, path := range options.DotNetReferences {
				ptLogF(t, "added .NET project reference: %s -> %s", name, path)
			}
		}

		// Set target framework if specified
		if options.DotNetTargetFramework != "" {
			err = setTargetFramework(csprojPath, options.DotNetTargetFramework)
			if err != nil {
				ptFatalF(t, "failed to set target framework: %s", err)
			}
			ptLogF(t, "set target framework to: %s", options.DotNetTargetFramework)
		}
	}

	// Set build configuration environment variable if specified
	if options.DotNetBuildConfig != "" {
		env["DOTNET_BUILD_CONFIGURATION"] = options.DotNetBuildConfig
		ptLogF(t, "set .NET build configuration to: %s", options.DotNetBuildConfig)
	}

	if err != nil {
		ptFatalF(t, "failed to create stack: %s", err)
		return nil
	}
	if !stackOptions.SkipDestroy {
		t.Cleanup(func() {
			t.Helper()

			if ptFailed(t) && skipDestroyOnFailure() {
				t.Log("Skipping destroy because PULUMITEST_SKIP_DESTROY_ON_FAILURE is set to 'true'.")
				writeDestroyScript(t, stack.Workspace().WorkDir(), stackName, env)
				return
			}

			t.Log("destroying stack, to skip this set PULUMITEST_SKIP_DESTROY_ON_FAILURE=true")
			_, err := stack.Destroy(pt.ctx)
			if err != nil {
				ptErrorF(t, "failed to destroy stack: %s", err)
			}
			err = stack.Workspace().RemoveStack(pt.ctx, stackName, optremove.Force())
			if err != nil {
				ptErrorF(t, "failed to remove stack: %s", err)
			}
		})
	}
	pt.currentStack = &stack
	return &stack
}

func writeDestroyScript(t PT, dir, stackName string, env map[string]string) {
	t.Helper()
	envPrefix := ""
	if passphrase, ok := env["PULUMI_CONFIG_PASSPHRASE"]; ok {
		envPrefix += fmt.Sprintf("export PULUMI_CONFIG_PASSPHRASE=%q\n", passphrase)
	}
	if backendUrl, ok := env["PULUMI_BACKEND_URL"]; ok {
		envPrefix += fmt.Sprintf("export PULUMI_BACKEND_URL=%q\n", backendUrl)
	}
	scriptContent := fmt.Sprintf(`#!/usr/bin/env bash
%s
cd "$(dirname "$0")" || exit
pulumi stack select %q
pulumi destroy --yes`, envPrefix, stackName)
	destroyScriptPath := filepath.Join(dir, "destroy.sh")
	os.WriteFile(destroyScriptPath, []byte(scriptContent), 0755)
	ptLogF(t, "Destroy can be run manually by running script at %q", destroyScriptPath)
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
