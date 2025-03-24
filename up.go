package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Up deploys the current stack.
// If an error is expected, use `CurrentStack().Up()` instead to be returned the error.
func (pt *PulumiTest) Up(t PT, opts ...optup.Option) auto.UpResult {
	t.Helper()

	t.Log("deploying")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !pt.options.DisableGrpcLog {
		pt.ClearGrpcLog(t)
	}
	result, err := pt.currentStack.Up(pt.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to deploy: %s", err)
	}
	return result
}
