package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ExportStack exports the current stack state.
func (a *PulumiTest) ExportStack() apitype.UntypedDeployment {
	a.t.Helper()

	a.t.Log("exporting stack")
	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	out, err := a.currentStack.Workspace().ExportStack(a.Context(), a.currentStack.Name())
	if err != nil {
		a.fatalf("failed to export stack: %s", err)
	}
	return out
}
