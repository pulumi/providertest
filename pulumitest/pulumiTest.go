package pulumitest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"gopkg.in/yaml.v3"
)

type PulumiTest struct {
	ctx          context.Context
	workingDir   string
	options      *opttest.Options
	currentStack *auto.Stack
}

// NewPulumiTest creates a new PulumiTest instance.
// By default it will:
// 1. Copy the source to a temporary directory.
// 2. Install dependencies.
// 3. Create a new stack called "test" with state stored to a local temporary directory and a fixed passphrase for encryption.
func NewPulumiTest(t PT, source string, opts ...opttest.Option) *PulumiTest {
	t.Helper()
	ctx := testContext(t)
	options := opttest.DefaultOptions()
	for _, opt := range opts {
		opt.Apply(options)
	}
	pt := &PulumiTest{
		ctx:        ctx,
		workingDir: source,
		options:    options,
	}
	if !options.TestInPlace {
		pt = pt.CopyToTempDir(t)
	} else {
		pulumiTestInit(t, pt, options)
	}
	return pt
}

func testContext(t PT) context.Context {
	t.Helper()
	var ctx context.Context
	var cancel context.CancelFunc
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	t.Cleanup(cancel)
	return ctx
}

// Perform the common initialization steps for a PulumiTest instance.
func pulumiTestInit(t PT, pt *PulumiTest, options *opttest.Options) {
	t.Helper()
	if !options.SkipInstall {
		// When PythonLinks are specified, we handle Python installation ourselves
		// instead of relying on `pulumi install`. This is because `pulumi install`
		// may fail on systems with PEP 668 restrictions (externally-managed Python).
		if len(options.PythonLinks) > 0 {
			pt.installPythonDependencies(t, options.PythonLinks)
		} else {
			pt.Install(t)
		}
	}
	if !options.SkipStackCreate {
		pt.NewStack(t, options.StackName, options.NewStackOpts...)
	}
}

// Deprecated: Use WorkingDir instead.
// Source returns the current working directory.
func (pt *PulumiTest) Source() string {
	return pt.workingDir
}

// WorkingDir returns the current working directory.
func (pt *PulumiTest) WorkingDir() string {
	return pt.workingDir
}

// Context returns the current context.Context instance used for automation API calls.
func (pt *PulumiTest) Context() context.Context {
	return pt.ctx
}

// CurrentStack returns the last stack that was created or nil if no stack has been created yet.
func (pt *PulumiTest) CurrentStack() *auto.Stack {
	return pt.currentStack
}

// pulumiYAMLRuntime represents the runtime section of Pulumi.yaml for validation
type pulumiYAMLRuntime struct {
	Name    string `yaml:"name"`
	Options struct {
		Virtualenv string `yaml:"virtualenv"`
		Toolchain  string `yaml:"toolchain"`
	} `yaml:"options"`
}

// pulumiYAML represents the minimal Pulumi.yaml structure needed for validation
type pulumiYAML struct {
	Runtime interface{} `yaml:"runtime"`
}

// validatePythonRuntimeConfig checks that Pulumi.yaml has the required runtime configuration
// for PythonLink to work correctly. Returns an error with helpful instructions if misconfigured.
func validatePythonRuntimeConfig(workingDir string) error {
	pulumiYamlPath := filepath.Join(workingDir, "Pulumi.yaml")
	data, err := os.ReadFile(pulumiYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read Pulumi.yaml: %w", err)
	}

	var config pulumiYAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse Pulumi.yaml: %w", err)
	}

	// Handle both string ("python") and object runtime configurations
	var runtimeConfig pulumiYAMLRuntime
	switch rt := config.Runtime.(type) {
	case string:
		// Simple format: "runtime: python" - missing required options
		if rt == "python" {
			return fmt.Errorf(`PythonLink requires Pulumi.yaml to specify the virtualenv location

Your Pulumi.yaml has:
  runtime: python

Please update it to:
  runtime:
    name: python
    options:
      toolchain: pip
      virtualenv: venv

This tells Pulumi to use the venv directory that PythonLink creates`)
		}
		return nil // Not a Python project
	case map[string]interface{}:
		// Object format - need to re-parse to get the full structure
		runtimeBytes, _ := yaml.Marshal(rt)
		if err := yaml.Unmarshal(runtimeBytes, &runtimeConfig); err != nil {
			return fmt.Errorf("failed to parse runtime config: %w", err)
		}
	default:
		return nil // Unknown format, let Pulumi handle it
	}

	// Check if it's a Python project
	if runtimeConfig.Name != "python" {
		return nil // Not a Python project
	}

	// Check for virtualenv configuration
	if runtimeConfig.Options.Virtualenv == "" {
		return fmt.Errorf(`PythonLink requires Pulumi.yaml to specify the virtualenv location

Your Pulumi.yaml has:
  runtime:
    name: python

Please add the virtualenv option:
  runtime:
    name: python
    options:
      toolchain: pip
      virtualenv: venv

This tells Pulumi to use the venv directory that PythonLink creates`)
	}

	return nil
}

// getPythonVenvPath returns the virtualenv path configured in Pulumi.yaml, or "venv" as default.
func getPythonVenvPath(workingDir string) string {
	pulumiYamlPath := filepath.Join(workingDir, "Pulumi.yaml")
	data, err := os.ReadFile(pulumiYamlPath)
	if err != nil {
		return "venv" // Default
	}

	var config pulumiYAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "venv" // Default
	}

	switch rt := config.Runtime.(type) {
	case map[string]interface{}:
		runtimeBytes, _ := yaml.Marshal(rt)
		var runtimeConfig pulumiYAMLRuntime
		if err := yaml.Unmarshal(runtimeBytes, &runtimeConfig); err == nil {
			if runtimeConfig.Options.Virtualenv != "" {
				return runtimeConfig.Options.Virtualenv
			}
		}
	}

	return "venv" // Default
}

// installPythonDependencies handles the complete Python dependency installation when
// PythonLinks are specified. This bypasses `pulumi install` entirely for Python projects
// to avoid PEP 668 issues where `pulumi install` tries to use system Python.
//
// The function:
// 1. Validates Pulumi.yaml has required runtime configuration
// 2. Creates a virtual environment at venv/
// 3. Upgrades pip in the venv
// 4. Installs local packages via pip install -e (PythonLinks)
// 5. Installs requirements.txt dependencies
// 6. Runs `pulumi plugin install` for Pulumi provider plugins
func (pt *PulumiTest) installPythonDependencies(t PT, pythonLinks []string) {
	t.Helper()

	// Validate Pulumi.yaml has required configuration
	if err := validatePythonRuntimeConfig(pt.workingDir); err != nil {
		ptFatalF(t, "%s", err)
	}

	ptLogF(t, "installing Python dependencies (bypassing pulumi install)")

	// Determine which Python interpreter to use. Try python3 first for better
	// compatibility with modern systems, then fall back to python.
	pythonCmd := "python"
	if _, err := exec.LookPath("python3"); err == nil {
		pythonCmd = "python3"
	}

	// Get the virtualenv path from Pulumi.yaml configuration
	venvDir := getPythonVenvPath(pt.workingDir)
	venvPath := filepath.Join(pt.workingDir, venvDir)

	// Step 1: Create venv if it doesn't exist
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
		cmd.Dir = pt.workingDir
		ptLogF(t, "creating virtual environment: %s", cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			ptFatalF(t, "failed to create virtual environment: %s\n%s", err, out)
		}
	}

	// Determine the pip and python paths inside the venv (platform-specific)
	var pipPath, venvPython string
	if _, err := os.Stat(filepath.Join(venvPath, "Scripts", "pip.exe")); err == nil {
		// Windows
		pipPath = filepath.Join(venvPath, "Scripts", "pip.exe")
		venvPython = filepath.Join(venvPath, "Scripts", "python.exe")
	} else {
		// Unix-like
		pipPath = filepath.Join(venvPath, "bin", "pip")
		venvPython = filepath.Join(venvPath, "bin", "python")
	}

	// Step 2: Upgrade pip in the venv (using venv's Python, not system Python)
	cmd := exec.Command(venvPython, "-m", "pip", "install", "--upgrade", "pip", "setuptools", "wheel")
	cmd.Dir = pt.workingDir
	ptLogF(t, "upgrading pip in venv: %s", cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		ptLogF(t, "warning: failed to upgrade pip (continuing anyway): %s\n%s", err, out)
		// Don't fail here - pip upgrade is nice-to-have but not required
	}

	// Step 3: Install local packages (PythonLinks) via editable install
	for _, pkgPath := range pythonLinks {
		absPath, err := filepath.Abs(pkgPath)
		if err != nil {
			ptFatalF(t, "failed to get absolute path for %s: %s", pkgPath, err)
		}
		cmd := exec.Command(pipPath, "install", "-e", absPath)
		cmd.Dir = pt.workingDir
		ptLogF(t, "installing local python package: %s", cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			ptFatalF(t, "failed to install python package %s: %s\n%s", pkgPath, err, out)
		}
	}

	// Step 4: Install requirements.txt if it exists
	requirementsPath := filepath.Join(pt.workingDir, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		cmd := exec.Command(pipPath, "install", "-r", requirementsPath)
		cmd.Dir = pt.workingDir
		ptLogF(t, "installing requirements.txt: %s", cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			ptFatalF(t, "failed to install requirements.txt: %s\n%s", err, out)
		}
	}

	// Step 5: Install Pulumi plugins
	cmd = exec.Command("pulumi", "plugin", "install")
	cmd.Dir = pt.workingDir
	ptLogF(t, "installing Pulumi plugins: %s", cmd)
	out, err = cmd.CombinedOutput()
	if err != nil {
		// Plugin install failure is often not fatal - the plugins may be
		// already installed or not needed for the test
		ptLogF(t, "warning: pulumi plugin install had issues: %s\n%s", err, out)
	}
}
