package pulumitest

import (
	"context"
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

type PulumiTest struct {
	t            *testing.T
	ctx          context.Context
	source       string
	options      *opttest.Options
	currentStack *auto.Stack
}

// NewPulumiTest creates a new PulumiTest instance.
// By default it will:
// 1. Copy the source to a temporary directory.
// 2. Install dependencies.
// 3. Create a new stack called "test" with state stored to a local temporary directory and a fixed passphrase for encryption.
func NewPulumiTest(t *testing.T, source string, opts ...opttest.Option) *PulumiTest {
	var ctx context.Context
	var cancel context.CancelFunc
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	t.Cleanup(cancel)
	options := opttest.DefaultOptions()
	for _, opt := range opts {
		opt.Apply(options)
	}
	pt := &PulumiTest{
		t:       t,
		ctx:     ctx,
		source:  source,
		options: options,
	}
	if !options.TestInPlace {
		return pt.CopyToTempDir()
	}
	if !options.SkipInstall {
		pt.Install()
	}
	if !options.SkipStackCreate {
		pt.NewStack(options.StackName)
	}
	return pt
}

// Source returns the current source directory.
func (a *PulumiTest) Source() string {
	return a.source
}

// T returns the current testing.T instance.
func (a *PulumiTest) T() *testing.T {
	return a.t
}

// Context returns the current context.Context instance used for automation API calls.
func (a *PulumiTest) Context() context.Context {
	return a.ctx
}

// CurrentStack returns the last stack that was created or nil if no stack has been created yet.
func (a *PulumiTest) CurrentStack() *auto.Stack {
	return a.currentStack
}
