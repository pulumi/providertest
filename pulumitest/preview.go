package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// Preview previews an update to the current stack.
func (a *PulumiTest) Preview(opts ...optpreview.Option) auto.PreviewResult {
	a.t.Helper()

	a.t.Log("previewing update")
	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog()
	}
	result, err := a.currentStack.Preview(a.ctx, opts...)
	if err != nil {
		a.fatalf("failed to preview update: %s", err)
	}
	return result
}
