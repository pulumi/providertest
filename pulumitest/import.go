package pulumitest

func (a *PulumiTest) Import(
	resourceType, resourceName, resourceID string, providerUrn string, args ...string,
) cmdOutput {
	a.t.Helper()
	arguments := []string{
		"import", resourceType, resourceName, resourceID, "--yes", "--protect=false", "-s", a.CurrentStack().Name(),
	}
	if providerUrn != "" {
		arguments = append(arguments, "--provider="+providerUrn)
	}
	arguments = append(arguments, args...)
	ret := a.execCmd(arguments...)
	if ret.ReturnCode != 0 {
		a.log(ret.Stdout)
		a.fatalf("failed to import resource %s: %s", resourceName, ret.Stderr)
	}

	return ret
}
