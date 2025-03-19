package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Destroy the current stack.
// If an error is expected, use `CurrentStack().Destroy()` instead to be returned the error.
func (pt *PulumiTest) Destroy(t PT, opts ...optdestroy.Option) auto.DestroyResult {
	t.Helper()

	t.Log("destroying")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	result, err := pt.currentStack.Destroy(pt.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to destroy: %s", err)
	}
	return result
}
