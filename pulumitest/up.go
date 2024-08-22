package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Up deploys the current stack.
func (a *PulumiTest) Up(t PT, opts ...optup.Option) auto.UpResult {
	t.Helper()

	t.Log("deploying")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	if !a.options.DisableGrpcLog {
		a.ClearGrpcLog(t)
	}
	result, err := a.currentStack.Up(a.ctx, opts...)
	if err != nil {
		ptFatalF(t, "failed to deploy: %s", err)
	}
	return result
}
