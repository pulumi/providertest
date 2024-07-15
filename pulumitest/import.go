package pulumitest

func (a *PulumiTest) Import(
	resourceType, resourceName, resourceID string, providerUrn string,
) {
	arguments := []string{
		"import", resourceType, resourceName, resourceID, "--yes", "--protect=false", "-s", a.CurrentStack().Name(),
	}
	if providerUrn != "" {
		arguments = append(arguments, "--provider="+providerUrn)
	}
	ret := a.execCmd(arguments...)
	if ret.ReturnCode != 0 {
		a.log(ret.Stdout)
		a.fatalf("failed to import resource %s: %s", resourceName, ret.Stderr)
	}
}
