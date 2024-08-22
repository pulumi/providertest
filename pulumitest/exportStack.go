package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ExportStack exports the current stack state.
func (a *PulumiTest) ExportStack(t PT) apitype.UntypedDeployment {
	t.Helper()

	t.Log("exporting stack")
	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	out, err := a.currentStack.Workspace().ExportStack(a.Context(), a.currentStack.Name())
	if err != nil {
		ptFatalF(t, "failed to export stack: %s", err)
	}
	return out
}
