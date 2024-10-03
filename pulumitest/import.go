package pulumitest

func (a *PulumiTest) Import(t PT, resourceType, resourceName, resourceID string, providerUrn string, args ...string) cmdOutput {
	t.Helper()
	arguments := []string{
		"import", resourceType, resourceName, resourceID, "--yes", "--protect=false", "-s", a.CurrentStack().Name(),
	}
	if providerUrn != "" {
		arguments = append(arguments, "--provider="+providerUrn)
	}
	arguments = append(arguments, args...)
	ret := a.execCmd(t, arguments...)
	if ret.ReturnCode != 0 {
		t.Log(ret.Stdout)
		ptFatalF(t, "failed to import resource %s: %s", resourceName, ret.Stderr)
	}

	return ret
}
