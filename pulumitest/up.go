package pulumitest

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Up deploys the current stack.
// If an error is expected, use `UpErr` instead to have the error returned.
func (pt *PulumiTest) Up(t PT, opts ...optup.Option) auto.UpResult {
	t.Helper()

	result, err := pt.UpErr(t, opts...)
	if err != nil {
		ptFatalF(t, "failed to deploy: %s", err)
	}
	return result
}

// UpErr deploys the current stack and returns any error instead of failing the test.
func (pt *PulumiTest) UpErr(t PT, opts ...optup.Option) (auto.UpResult, error) {
	t.Helper()

	t.Log("deploying")
	if pt.currentStack == nil {
		return auto.UpResult{}, fmt.Errorf("no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	var result auto.UpResult
	err := pt.withProviders(t, pt.currentStack, func() error {
		var upErr error
		result, upErr = pt.currentStack.Up(pt.ctx, opts...)
		return upErr
	})
	return result, err
}
