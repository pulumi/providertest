package pulumitest

import (
	"github.com/pulumi/providertest/pulumitest/optnewstack"
)

// InstallStack installs packages, and creates a new stack.
func (pt *PulumiTest) InstallStack(stackName string, opts ...optnewstack.NewStackOpt) *PulumiTest {
	pt.t.Helper()

	pt.Install()
	pt.NewStack(stackName, opts...)
	return pt
}
