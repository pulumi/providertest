package pulumitest

// Import performs a `pulumi import` operation on the current stack.
// The resource type, name, and ID are required. The provider URN is optional.
func (pt *PulumiTest) Import(t PT, resourceType, resourceName, resourceID string, providerUrn string, args ...string) cmdOutput {
	t.Helper()
	arguments := []string{
		"import", resourceType, resourceName, resourceID, "--yes", "--protect=false", "-s", pt.CurrentStack().Name(),
	}
	if providerUrn != "" {
		arguments = append(arguments, "--provider="+providerUrn)
	}
	arguments = append(arguments, args...)
	ret := pt.execCmd(t, arguments...)
	if ret.ReturnCode != 0 {
		t.Log(ret.Stdout)
		ptFatalF(t, "failed to import resource %s: %s", resourceName, ret.Stderr)
	}

	return ret
}
