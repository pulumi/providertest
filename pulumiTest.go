package pulumitest

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumitest/opttest"
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
