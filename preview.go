package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// Preview previews an update to the current stack.
// If an error is expected, use `CurrentStack().Preview()` instead to be returned the error.
func (pt *PulumiTest) Preview(t PT, opts ...optpreview.Option) auto.PreviewResult {
	t.Helper()

	t.Log("previewing update")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	result, err := pt.currentStack.Preview(pt.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to preview update: %s", err)
	}
	return result
}
