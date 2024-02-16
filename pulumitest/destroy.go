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
		a.t.Fatal("no current stack")
	}
	a.ClearGrpcLog()
	result, err := a.currentStack.Destroy(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to destroy: %s", err)
	}
	return result
}
