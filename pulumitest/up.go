package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Up deploys the current stack.
func (a *PulumiTest) Up(opts ...optup.Option) auto.UpResult {
	a.t.Helper()

	a.t.Log("deploying")
	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog()
	}
	result, err := a.currentStack.Up(a.ctx, opts...)
	if err != nil {
		a.fatalf("failed to deploy: %s", err)
	}
	return result
}
