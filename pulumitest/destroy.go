package pulumitest

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Destroy the current stack.
// If an error is expected, use `DestroyErr` instead to have the error returned.
func (pt *PulumiTest) Destroy(t PT, opts ...optdestroy.Option) auto.DestroyResult {
	t.Helper()

	result, err := pt.DestroyErr(t, opts...)
	if err != nil {
		ptFatalF(t, "failed to destroy: %s", err)
	}
	return result
}

// DestroyErr destroys the current stack and returns any error instead of failing the test.
func (pt *PulumiTest) DestroyErr(t PT, opts ...optdestroy.Option) (auto.DestroyResult, error) {
	t.Helper()

	t.Log("destroying")
	if pt.currentStack == nil {
		return auto.DestroyResult{}, fmt.Errorf("no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	var result auto.DestroyResult
	err := pt.withProviders(t, pt.currentStack, func() error {
		var destroyErr error
		result, destroyErr = pt.currentStack.Destroy(pt.ctx, opts...)
		return destroyErr
	})
	return result, err
}
