package pulumitest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// InstallStack installs packages, and creates a new stack.
func (pt *PulumiTest) InstallStack(stackName string, opts ...auto.LocalWorkspaceOption) *PulumiTest {
	pt.t.Helper()

	pt.Install()
	pt.NewStack(stackName, opts...)
	return pt
}
