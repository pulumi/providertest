package autotest

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

type AutoTest struct {
	t            *testing.T
	ctx          context.Context
	source       string
	envBuilder   *EnvBuilder
	currentStack *auto.Stack
}

func NewAutoTest(t *testing.T, source string) *AutoTest {
	deadline, ok := t.Deadline()
	var ctx context.Context
	if ok {
		ctx, _ = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	return &AutoTest{
		t:          t,
		ctx:        ctx,
		source:     source,
		envBuilder: NewEnvBuilder(t),
	}
}

func (a *AutoTest) Source() string {
	return a.source
}

func (a *AutoTest) T() *testing.T {
	return a.t
}

func (a *AutoTest) Context() context.Context {
	return a.ctx
}

func (a *AutoTest) Env() *EnvBuilder {
	return a.envBuilder
}

func (a *AutoTest) WithSource(source string) *AutoTest {
	a.t.Helper()
	a.source = source
	return a
}

func (a *AutoTest) CurrentStack() *auto.Stack {
	return a.currentStack
}
