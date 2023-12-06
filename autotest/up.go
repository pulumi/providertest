package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Up deploys the current stack.
func (a *AutoTest) Up(opts ...optup.Option) auto.UpResult {
	a.t.Helper()

	a.t.Log("deploying")
	if a.currentStack == nil {
		a.t.Fatal("no current stack")
	}
	result, err := a.currentStack.Up(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to deploy: %s", err)
	}
	return result
}
