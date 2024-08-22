package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
)

// Refresh refreshes the current stack.
func (a *PulumiTest) Refresh(t PT, opts ...optrefresh.Option) auto.RefreshResult {
	t.Helper()

	t.Log("refreshing")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog(t)
	}
	result, err := a.currentStack.Refresh(a.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to refresh: %s", err)
	}
	return result
}
