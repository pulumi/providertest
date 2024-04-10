package pulumitest

import (
	"context"
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

type PulumiTest struct {
	t            PT
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
func NewPulumiTest(t PT, source string, opts ...opttest.Option) *PulumiTest {
	t.Helper()
	ctx := testContext(t)
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
		pt = pt.CopyToTempDir()
	} else {
		pulumiTestInit(pt, options)
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
func pulumiTestInit(pt *PulumiTest, options *opttest.Options) {
	pt.t.Helper()
	if !options.SkipInstall {
		pt.Install()
	}
	if !options.SkipStackCreate {
		pt.NewStack(options.StackName, options.NewStackOpts...)
	}
}

// Source returns the current source directory.
func (a *PulumiTest) Source() string {
	return a.source
}

// PT returns the current PT instance.
func (a *PulumiTest) PT() PT {
	return a.t
}

// Please use [PulumiTest.PT] instead.
// Deprecated.
func (a *PulumiTest) T() *testing.T {
	r, ok := a.PT().(*testing.T)
	contract.Assertf(ok, "Testing with something other than *testing.T; Please use PulumiTest.PT()")
	return r
}

// Context returns the current context.Context instance used for automation API calls.
func (a *PulumiTest) Context() context.Context {
	return a.ctx
}

// CurrentStack returns the last stack that was created or nil if no stack has been created yet.
func (a *PulumiTest) CurrentStack() *auto.Stack {
	return a.currentStack
}
