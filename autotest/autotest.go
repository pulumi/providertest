package autotest

import (
	"context"
	"testing"
)

type AutoTest struct {
	t      *testing.T
	ctx    context.Context
	source string
}

func NewAutoTest(t *testing.T, source string) *AutoTest {
	deadline, ok := t.Deadline()
	var ctx context.Context
	if ok {
		ctx, _ = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx = context.Background()
	}
	return &AutoTest{t: t, ctx: ctx, source: source}
}

func (a *AutoTest) Source() string {
	return a.source
}

func (a *AutoTest) T() *testing.T {
	return a.t
}

func (a *AutoTest) Ctx() context.Context {
	return a.ctx
}

func (a *AutoTest) WithSource(source string) *AutoTest {
	return &AutoTest{t: a.t, ctx: a.ctx, source: source}
}
