package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
)

// Refresh refreshes the current stack.
// If an error is expected, use `CurrentStack().Refresh()` instead to be returned the error.
func (pt *PulumiTest) Refresh(t PT, opts ...optrefresh.Option) auto.RefreshResult {
	t.Helper()

	t.Log("refreshing")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	var result auto.RefreshResult
	err := pt.withProviders(t, func() error {
		var refreshErr error
		result, refreshErr = pt.currentStack.Refresh(pt.ctx, opts...)
		return refreshErr
	})
	if err != nil {
		ptFatalF(t, "failed to refresh: %s", err)
	}
	return result
}
