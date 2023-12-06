package autotest

// Init copies the program to a temporary directory, restores packages, and creates a new stack.
func (a *AutoTest) Init(stackName string) *AutoTest {
	a.t.Helper()

	copy := a.CopyToTempDir()
	copy.Install()
	copy.NewStack(stackName)
	return copy
}
