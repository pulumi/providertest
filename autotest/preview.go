package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

func (a *AutoTest) Preview(opts ...optpreview.Option) auto.PreviewResult {
	a.t.Helper()

	a.t.Log("previewing update")
	result, err := a.currentStack.Preview(a.ctx, opts...)
	if err != nil {
		a.t.Fatalf("failed to preview update: %s", err)
	}
	return result
}
