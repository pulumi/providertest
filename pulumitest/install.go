package pulumitest

import (
	"os/exec"
)

// Install installs packages and plugins for a given directory by running `pulumi install`.
func (pt *PulumiTest) Install(t PT) string {
	t.Helper()

	t.Log("installing packages and plugins")
	cmd := exec.Command("pulumi", "install")
	cmd.Dir = pt.workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		ptFatalF(t, "failed to install packages and plugins: %s\n%s", err, out)
	}
	return string(out)
}
