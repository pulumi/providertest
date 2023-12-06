package autotest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// Init copies the program to a temporary directory, restores packages, and creates a new stack.
func (a *AutoTest) Init(stackName string, opts ...auto.LocalWorkspaceOption) *AutoTest {
	a.t.Helper()

	copy := a.CopyToTempDir()
	copy.Install()
	copy.NewStack(stackName, opts...)
	return copy
}
