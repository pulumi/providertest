package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
)

// Refresh refreshes the current stack.
func (a *PulumiTest) Refresh(opts ...optrefresh.Option) auto.RefreshResult {
	a.t.Helper()

	a.t.Log("refreshing")
	if a.currentStack == nil {
		a.t.Fatal("no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog()
	}
	result, err := a.currentStack.Refresh(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to refresh: %s", err)
	}
	return result
}
