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
		a.t.Fatal("no current stack")
	}
	a.ClearGrpcLog()
	result, err := a.currentStack.Preview(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to preview update: %s", err)
	}
	return result
}
