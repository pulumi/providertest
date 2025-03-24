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
	result, err := pt.currentStack.Refresh(pt.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to refresh: %s", err)
	}
	return result
}
