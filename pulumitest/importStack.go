package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ImportStack imports the given stack state into the test's current stack.
func (a *PulumiTest) ImportStack(source apitype.UntypedDeployment) {
	a.t.Helper()

	a.t.Log("importing stack")
	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	err := a.currentStack.Workspace().ImportStack(a.Context(), a.currentStack.Name(), source)
	if err != nil {
		a.fatalf("failed to import stack: %s", err)
	}
}
