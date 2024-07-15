package pulumitest


func (a *PulumiTest) StateDelete(resourceURN string) {
	arguments := []string{
		"state", "delete", resourceURN, "--yes", "-s", a.CurrentStack().Name(),
	}
	ret := a.execCmd(arguments...)
	if ret.ReturnCode != 0 {
		a.log(ret.Stdout)
		a.fatalf("Failed to delete from state %s: %s", resourceURN, ret.Stderr)
	}
}
