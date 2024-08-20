package pulumitest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/providertest/pulumiyaml"
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

// NewPulumiYamlTest creates a new PulumiYamlTest instance.
func NewInlinePulumiTest(t PT, program *pulumiyaml.Program, opts ...opttest.Option) *PulumiTest {
	t.Helper()
	ctx := testContext(t)
	options := opttest.DefaultOptions()
	for _, opt := range opts {
		opt.Apply(options)
	}
	if program.Runtime == "" {
		program.Runtime = "yaml"
	}
	if program.Name == "" {
		program.Name = "test"
	}
	workingDir := tempDirWithoutCleanupOnFailedTest(t, "yamlProgramDir")
	pt := &PulumiTest{
		ctx:        ctx,
		workingDir: workingDir,
		options:    options,
	}
	yamlBytes, err := yaml.Marshal(program)
	if err != nil {
		ptFatalF(t, "failed to marshal yaml program: %v", err)
	}
	yamlPath := filepath.Join(workingDir, "Pulumi.yaml")
	err = os.WriteFile(yamlPath, yamlBytes, 0644)
	if err != nil {
		ptFatalF(t, "failed to write yaml program: %v", err)
	}
	pulumiTestInit(t, pt, options)
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
		pt.Install(t)
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
