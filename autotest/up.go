package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

func (a *AutoTest) Up(stack *auto.Stack, opts ...optup.Option) auto.UpResult {
	a.t.Helper()

	a.t.Log("deploying")
	result, err := stack.Up(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to deploy: %s", err)
	}
	return result
}
