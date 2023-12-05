package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func (a *AutoTest) ImportStack(source apitype.UntypedDeployment) {
	a.t.Helper()

	a.t.Log("importing stack")
	err := a.currentStack.Workspace().ImportStack(a.Context(), a.currentStack.Name(), source)
	if err != nil {
		a.t.Fatalf("failed to import stack: %s", err)
	}
}
