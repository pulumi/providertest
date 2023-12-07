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
	options      []opttest.Option
	currentStack *auto.Stack
}

func NewPulumiTest(t *testing.T, source string, opts ...opttest.Option) *PulumiTest {
	return NewPulumiTestInPlace(t, source, opts...).CopyToTempDir()
}

func NewPulumiTestInPlace(t *testing.T, source string, opts ...opttest.Option) *PulumiTest {
	var ctx context.Context
	var cancel context.CancelFunc
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	t.Cleanup(cancel)
	return &PulumiTest{
		t:       t,
		ctx:     ctx,
		source:  source,
		options: opts,
	}
}

func (a *PulumiTest) Source() string {
	return a.source
}

func (a *PulumiTest) T() *testing.T {
	return a.t
}

func (a *PulumiTest) Context() context.Context {
	return a.ctx
}

func (a *PulumiTest) WithSource(source string) *PulumiTest {
	a.t.Helper()
	a.source = source
	return a
}

func (a *PulumiTest) CurrentStack() *auto.Stack {
	return a.currentStack
}
