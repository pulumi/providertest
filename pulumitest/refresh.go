package pulumitest

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
)

// Refresh refreshes the current stack.
// If an error is expected, use `RefreshErr` instead to have the error returned.
func (pt *PulumiTest) Refresh(t PT, opts ...optrefresh.Option) auto.RefreshResult {
	t.Helper()

	result, err := pt.RefreshErr(t, opts...)
	if err != nil {
		ptFatalF(t, "failed to refresh: %s", err)
	}
	return result
}

// RefreshErr refreshes the current stack and returns any error instead of failing the test.
func (pt *PulumiTest) RefreshErr(t PT, opts ...optrefresh.Option) (auto.RefreshResult, error) {
	t.Helper()

	t.Log("refreshing")
	if pt.currentStack == nil {
		return auto.RefreshResult{}, fmt.Errorf("no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	var result auto.RefreshResult
	err := pt.withProviders(t, pt.currentStack, func() error {
		var refreshErr error
		result, refreshErr = pt.currentStack.Refresh(pt.ctx, opts...)
		return refreshErr
	})
	return result, err
}
