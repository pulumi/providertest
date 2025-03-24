package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ImportStack imports the given stack state into the test's current stack.
func (pt *PulumiTest) ImportStack(t PT, source apitype.UntypedDeployment) {
	t.Helper()

	t.Log("importing stack")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	err := pt.currentStack.Workspace().ImportStack(pt.Context(), pt.currentStack.Name(), source)
	if err != nil {
		ptFatalF(t, "failed to import stack: %s", err)
	}
}
