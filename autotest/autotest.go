package autotest

import (
	"context"
	"testing"
)

type AutoTest struct {
	t          *testing.T
	ctx        context.Context
	source     string
	providers  map[string]ProviderFactory
	envBuilder *EnvBuilder
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
		providers:  map[string]ProviderFactory{},
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
