package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Up deploys the current stack.
func (a *PulumiTest) Destroy(t PT, opts ...optdestroy.Option) auto.DestroyResult {
	t.Helper()

	t.Log("destroying")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog(t)
	}
	result, err := a.currentStack.Destroy(a.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to destroy: %s", err)
	}
	return result
}
