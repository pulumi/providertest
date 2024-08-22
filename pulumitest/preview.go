package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// Preview previews an update to the current stack.
func (a *PulumiTest) Preview(t PT, opts ...optpreview.Option) auto.PreviewResult {
	t.Helper()

	t.Log("previewing update")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog(t)
	}
	result, err := a.currentStack.Preview(a.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to preview update: %s", err)
	}
	return result
}
