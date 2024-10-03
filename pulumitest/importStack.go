package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ImportStack imports the given stack state into the test's current stack.
func (a *PulumiTest) ImportStack(t PT, source apitype.UntypedDeployment) {
	t.Helper()

	t.Log("importing stack")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	err := a.currentStack.Workspace().ImportStack(a.Context(), a.currentStack.Name(), source)
	if err != nil {
		ptFatalF(t, "failed to import stack: %s", err)
	}
}
