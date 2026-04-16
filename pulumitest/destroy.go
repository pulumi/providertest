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
	var result auto.DestroyResult
	err := pt.withProviders(t, func() error {
		var destroyErr error
		result, destroyErr = pt.currentStack.Destroy(pt.ctx, opts...)
		return destroyErr
	})
	if err != nil {
		ptFatalF(t, "failed to destroy: %s", err)
	}
	return result
}
