package pulumitest

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// Preview previews an update to the current stack.
// If an error is expected, use `PreviewErr` instead to have the error returned.
func (pt *PulumiTest) Preview(t PT, opts ...optpreview.Option) auto.PreviewResult {
	t.Helper()

	result, err := pt.PreviewErr(t, opts...)
	if err != nil {
		ptFatalF(t, "failed to preview update: %s", err)
	}
	return result
}

// PreviewErr previews an update and returns any error instead of failing the test.
func (pt *PulumiTest) PreviewErr(t PT, opts ...optpreview.Option) (auto.PreviewResult, error) {
	t.Helper()

	t.Log("previewing update")
	if pt.currentStack == nil {
		return auto.PreviewResult{}, fmt.Errorf("no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	var result auto.PreviewResult
	err := pt.withProviders(t, pt.currentStack, func() error {
		var previewErr error
		result, previewErr = pt.currentStack.Preview(pt.ctx, opts...)
		return previewErr
	})
	return result, err
}
