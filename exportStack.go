package pulumitest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// ExportStack exports the current stack state.
func (pt *PulumiTest) ExportStack(t PT) apitype.UntypedDeployment {
	t.Helper()

	t.Log("exporting stack")
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	out, err := pt.currentStack.Workspace().ExportStack(pt.Context(), pt.currentStack.Name())
	if err != nil {
		ptFatalF(t, "failed to export stack: %s", err)
	}
	return out
}
