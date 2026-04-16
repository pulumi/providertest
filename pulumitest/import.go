package pulumitest

import "fmt"

// Import performs a `pulumi import` operation on the current stack.
// The resource type, name, and ID are required. The provider URN is optional.
func (pt *PulumiTest) Import(t PT, resourceType, resourceName, resourceID string, providerUrn string, args ...string) cmdOutput {
	t.Helper()
	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
		return cmdOutput{}
	}
	arguments := []string{
		"import", resourceType, resourceName, resourceID, "--yes", "--protect=false", "-s", pt.currentStack.Name(),
	}
	if providerUrn != "" {
		arguments = append(arguments, "--provider="+providerUrn)
	}
	arguments = append(arguments, args...)
	var ret cmdOutput
	err := pt.withProviders(t, pt.currentStack, func() error {
		ret = pt.execCmd(t, arguments...)
		if ret.ReturnCode != 0 {
			return fmt.Errorf("failed to import resource %s: %s", resourceName, ret.Stderr)
		}
		return nil
	})
	if err != nil {
		if ret.ReturnCode != 0 && ret.Stdout != "" {
			t.Log(ret.Stdout)
		}
		ptFatalF(t, "%s", err)
	}

	return ret
}
