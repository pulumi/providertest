package pulumitest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// Init copies the program to a temporary directory, restores packages, and creates a new stack.
func (a *PulumiTest) Init(stackName string, opts ...auto.LocalWorkspaceOption) *PulumiTest {
	a.t.Helper()

	copy := a.CopyToTempDir()
	copy.Install()
	copy.NewStack(stackName, opts...)
	return copy
}
