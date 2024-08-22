package pulumitest

import (
	"github.com/pulumi/providertest/pulumitest/optnewstack"
)

// InstallStack installs packages, and creates a new stack.
func (pt *PulumiTest) InstallStack(t PT, stackName string, opts ...optnewstack.NewStackOpt) *PulumiTest {
	t.Helper()

	pt.Install(t)
	pt.NewStack(t, stackName, opts...)
	return pt
}
