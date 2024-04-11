package pulumitest

import (
	"os/exec"
)

// Install installs packages and plugins for a given directory by running `pulumi install`.
func (a *PulumiTest) Install() string {
	a.t.Helper()

	a.t.Log("installing packages and plugins")
	cmd := exec.Command("pulumi", "install")
	cmd.Dir = a.source
	out, err := cmd.CombinedOutput()
	if err != nil {
		a.fatalf("failed to install packages and plugins: %s\n%s", err, out)
	}
	return string(out)
}
