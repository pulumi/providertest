package pulumitest

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremove"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// NewStack creates a new stack, ensure it's cleaned up after the test is done.
// If no stack name is provided, a random one will be generated.
func (pt *PulumiTest) NewStack(stackName string, opts ...auto.LocalWorkspaceOption) *auto.Stack {
	pt.t.Helper()

	if stackName == "" {
		stackName = randomStackName(pt.source)
	}

	options := opttest.NewOptions()
	for _, o := range pt.options {
		o.Apply(options)
	}

	// Set default stack opts. These can be overridden by the caller.
	env := map[string]string{}

	if options.ConfigPassphrase != "" {
		env["PULUMI_CONFIG_PASSPHRASE"] = options.ConfigPassphrase
	}

	if !options.UseAmbientBackend {
		backendFolder := pt.t.TempDir()
		env["PULUMI_BACKEND_URL"] = "file://" + backendFolder
	}

	if len(options.ProviderFactories) > 0 {
		pt.t.Log("starting providers")
		providerPorts, cancel, err := providers.StartProviders(pt.ctx, options.ProviderFactories)
		if err != nil {
			pt.t.Fatalf("failed to start providers: %v", err)
		}
		pt.t.Cleanup(func() {
			cancel()
		})
		env["PULUMI_DEBUG_PROVIDERS"] = providers.GetDebugProvidersEnv(providerPorts)
	}

	stackOpts := []auto.LocalWorkspaceOption{
		auto.EnvVars(env),
	}
	stackOpts = append(stackOpts, options.ExtraWorkspaceOptions...)
	stackOpts = append(stackOpts, opts...)

	stack, err := auto.NewStackLocalSource(pt.ctx, stackName, pt.source, stackOpts...)

	if options.ProviderPluginPaths != nil && len(options.ProviderPluginPaths) > 0 {
		projectSettings, err := stack.Workspace().ProjectSettings(pt.ctx)
		if err != nil {
			pt.t.Fatalf("failed to get project settings: %s", err)
		}
		var plugins workspace.Plugins
		if projectSettings.Plugins != nil {
			plugins = *projectSettings.Plugins
		}
		providers := plugins.Providers
		// Sort the provider plugin paths to ensure a consistent order.
		providerPluginNames := make([]string, 0, len(options.ProviderPluginPaths))
		for name := range options.ProviderPluginPaths {
			providerPluginNames = append(providerPluginNames, name)
		}
		sort.Strings(providerPluginNames)
		for _, name := range providerPluginNames {
			path := options.ProviderPluginPaths[name]
			found := false
			for _, provider := range providers {
				if provider.Name == name {
					provider.Path = path
					found = true
					break
				}
			}
			if !found {
				providers = append(providers, workspace.PluginOptions{
					Name: name,
					Path: path,
				})
			}
		}
		plugins.Providers = providers
		projectSettings.Plugins = &plugins
		err = stack.Workspace().SaveProjectSettings(pt.ctx, projectSettings)
		if err != nil {
			pt.t.Fatalf("failed to save project settings: %s", err)
		}
	}

	if err != nil {
		pt.t.Fatalf("failed to create stack: %s", err)
		return nil
	}
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
