package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func (a *AutoTest) ExportStack() apitype.UntypedDeployment {
	a.t.Helper()

	a.t.Log("exporting stack")
	out, err := a.currentStack.Workspace().ExportStack(a.Context(), a.currentStack.Name())
	if err != nil {
		a.t.Fatalf("failed to export stack: %s", err)
	}
	return out
}
