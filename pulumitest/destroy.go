package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Up deploys the current stack.
func (a *PulumiTest) Destroy(opts ...optdestroy.Option) auto.DestroyResult {
	a.t.Helper()

	a.t.Log("destroying")
	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog()
	}
	result, err := a.currentStack.Destroy(a.ctx, opts...)
	if err != nil {
		a.fatalf("failed to destroy: %s", err)
	}
	return result
}
