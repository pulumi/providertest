package pulumitest

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

func resolveJavaMavenDependencyPath(path, artifactID string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		base := filepath.Base(absPath)
		if !strings.HasSuffix(base, ".jar") {
			return "", fmt.Errorf("dependency path %s must point to a .jar file or directory containing one", absPath)
		}
		if strings.HasSuffix(base, "-sources.jar") || strings.HasSuffix(base, "-javadoc.jar") {
			return "", fmt.Errorf("dependency path %s must point to a runtime JAR, not a sources or javadoc archive", absPath)
		}
		return absPath, nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Maven dependency directory %s: %w", absPath, err)
	}

	var exactMatches []string
	var jarMatches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jar") {
			continue
		}
		if strings.HasSuffix(name, "-sources.jar") || strings.HasSuffix(name, "-javadoc.jar") {
			continue
		}

		fullPath := filepath.Join(absPath, name)
		jarMatches = append(jarMatches, fullPath)
		if strings.HasPrefix(name, artifactID+"-") || name == artifactID+".jar" {
			exactMatches = append(exactMatches, fullPath)
		}
	}

	sort.Strings(exactMatches)
	sort.Strings(jarMatches)

	switch {
	case len(exactMatches) == 1:
		return exactMatches[0], nil
	case len(exactMatches) > 1:
		return "", fmt.Errorf("multiple JARs matching artifact %s found in %s", artifactID, absPath)
	case len(jarMatches) == 1:
		return jarMatches[0], nil
	case len(jarMatches) > 1:
		return "", fmt.Errorf("multiple JARs found in %s; provide a specific JAR path", absPath)
	default:
		return "", fmt.Errorf("no JAR found in %s; provide a specific JAR path", absPath)
	}
}

func (pt *PulumiTest) workspaceEnv(t PT) map[string]string {
	t.Helper()

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

	// Apply custom env last to allow overriding any of the above.
	for k, v := range options.CustomEnv {
		env[k] = v
	}

	// Note: These environment variables are not standard Maven variables.
	// Their support depends on Pulumi's Java runtime implementation.
	if options.JavaMavenProfile != "" {
		env["MAVEN_ACTIVE_PROFILES"] = options.JavaMavenProfile
		ptLogF(t, "setting Maven active profile: %s", options.JavaMavenProfile)
	}

	if options.JavaMavenSettings != "" {
		absSettingsPath, err := filepath.Abs(options.JavaMavenSettings)
		if err != nil {
			ptFatalF(t, "failed to get absolute path for Maven settings: %s", err)
		}
		if _, err := os.Stat(absSettingsPath); err != nil {
			ptFatalF(t, "Maven settings file not found at %s: %s", absSettingsPath, err)
		}
		env["MAVEN_SETTINGS"] = absSettingsPath
		ptLogF(t, "setting Maven settings file: %s", absSettingsPath)
	}

	return env
}

func (pt *PulumiTest) prepareJavaWorkspace(t PT) {
	t.Helper()

	if pt.javaPrepared {
		return
	}

	options := pt.options
	if options.JavaTargetVersion == "" && len(options.JavaMavenDependencies) == 0 {
		pt.javaPrepared = true
		return
	}

	pomPath, err := findPomFile(pt.workingDir)
	if err != nil {
		ptFatalF(t, "Java options specified but pom.xml not found in %s: %s",
			pt.workingDir, err)
	}

	if options.JavaTargetVersion != "" {
		ptLogF(t, "setting Java target version to %s", options.JavaTargetVersion)
		if err := setJavaVersion(pomPath, options.JavaTargetVersion); err != nil {
			ptFatalF(t, "failed to set Java version in pom.xml: %s", err)
		}
	}

	if len(options.JavaMavenDependencies) > 0 {
		orderedDeps := make([]string, 0, len(options.JavaMavenDependencies))
		for depKey := range options.JavaMavenDependencies {
			orderedDeps = append(orderedDeps, depKey)
		}
		sort.Strings(orderedDeps)

		for _, depKey := range orderedDeps {
			dep := options.JavaMavenDependencies[depKey]
			resolvedPath, err := resolveJavaMavenDependencyPath(dep.Path, dep.ArtifactID)
			if err != nil {
				ptFatalF(t, "failed to resolve Maven dependency for %s:%s at %s: %s",
					dep.GroupID, dep.ArtifactID, dep.Path, err)
			}

			version := dep.Version
			if version == "" {
				version = "0.0.0-dev"
			}

			ptLogF(t, "adding Maven dependency %s:%s@%s with path %s", dep.GroupID, dep.ArtifactID, version, resolvedPath)
			if err := addOrUpdateDependency(pomPath, dep.GroupID, dep.ArtifactID, version, resolvedPath); err != nil {
				ptFatalF(t, "failed to add Maven dependency: %s", err)
			}
		}
	}

	pt.javaPrepared = true
}

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
	pt.prepareJavaWorkspace(t)
	env := pt.workspaceEnv(t)

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

	if len(options.YarnLinks) > 0 {
		for _, pkg := range options.YarnLinks {
			cmd := exec.Command("yarn", "link", pkg)
			cmd.Dir = pt.workingDir
			ptLogF(t, "linking yarn package: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				ptFatalF(t, "failed to link yarn package %s: %s\n%s", pkg, err, out)
			}
		}
	} else {
		projectSettings, err := stack.Workspace().ProjectSettings(pt.ctx)
		if err != nil {
			ptFatalF(t, "failed to get project settings: %s", err)
		}
		if projectSettings.Runtime.Name() == "nodejs" && len(options.YarnLinks) == 0 {
			if options.RequireYarnLinks == nil {
				ptLogF(t, "WARNING: YarnLinks were not set, but project runtime is nodejs. Module under test may not be used. Pass RequireYarnLinks(false) to silence this warning.")
			} else if *options.RequireYarnLinks {
				ptFatalF(t, "module under test may not be used: YarnLinks were not set, but project runtime is nodejs and RequireYarnLinks is true.")
			}
			// else: User decided to silence the warning explicitly by passing RequireYarnLinks(false)
		}
	}

	if len(options.PythonLinks) > 0 {
		// Determine which Python interpreter to use. Try python3 first for better
		// compatibility with modern systems, then fall back to python.
		pythonCmd := "python"
		if _, err := exec.LookPath("python3"); err == nil {
			pythonCmd = "python3"
		}

		for _, pkgPath := range options.PythonLinks {
			absPath, err := filepath.Abs(pkgPath)
			if err != nil {
				ptFatalF(t, "failed to get absolute path for %s: %s", pkgPath, err)
			}
			cmd := exec.Command(pythonCmd, "-m", "pip", "install", "-e", absPath)
			cmd.Dir = pt.workingDir
			ptLogF(t, "installing python package: %s", cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				ptFatalF(t, "failed to install python package %s: %s\n%s", pkgPath, err, out)
			}
		}
	}

	if len(options.GoModReplacements) > 0 {
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
	if len(options.DotNetReferences) > 0 {
		// Find the .csproj file in the working directory
		csprojPath, err := findCsprojFile(pt.workingDir)
		if err != nil {
			ptFatalF(t, "failed to find .csproj file: %s", err)
		}
		ptLogF(t, "found .csproj file: %s", csprojPath)

		// Add project references
		err = addProjectReferences(csprojPath, options.DotNetReferences)
		if err != nil {
			ptFatalF(t, "failed to add .NET project references: %s", err)
		}

		// Log the references that were added
		for name, path := range options.DotNetReferences {
			ptLogF(t, "added .NET project reference: %s -> %s", name, path)
		}
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
			err := pt.withProviders(t, &stack, func() error {
				_, destroyErr := stack.Destroy(pt.ctx)
				return destroyErr
			})
			if err != nil {
				if errors.Is(err, errStartProviders) {
					ptErrorF(t, "failed to start providers for cleanup destroy of stack %q; leaving stack state for manual cleanup: %s", stackName, err)
				} else {
					ptErrorF(t, "failed to destroy stack %q during cleanup; leaving stack state for manual cleanup: %s", stackName, err)
				}
				writeDestroyScript(t, stack.Workspace().WorkDir(), stackName, env)
				return
			}
			err = stack.Workspace().RemoveStack(pt.ctx, stackName, optremove.Force())
			if err != nil {
				ptErrorF(t, "failed to remove stack: %s", err)
			}
		})
	}
	pt.currentStack = &stack
	if !options.DisablePulumiVersionLog {
		pt.logPulumiVersionInfo(t)
	}
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
	if err := os.WriteFile(destroyScriptPath, []byte(scriptContent), 0755); err != nil {
		ptLogF(t, "failed to write destroy script: %v", err)
		return
	}
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
