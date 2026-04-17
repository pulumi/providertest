package pulumitest

import (
	"fmt"
	"strings"
)

func formatCommandDiagnostics(stdout, stderr string) string {
	parts := make([]string, 0, 2)
	if trimmed := strings.TrimSpace(stdout); trimmed != "" {
		parts = append(parts, "stdout: "+trimmed)
	}
	if trimmed := strings.TrimSpace(stderr); trimmed != "" {
		parts = append(parts, "stderr: "+trimmed)
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, "; ") + ")"
}

func (pt *PulumiTest) logPulumiVersionInfo(t PT) {
	t.Helper()
	workspace := pt.currentStack.Workspace()
	cmd := workspace.PulumiCommand()
	workdir := workspace.WorkDir()
	env := make([]string, 0, len(workspace.GetEnvVars()))
	for k, v := range workspace.GetEnvVars() {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	if stdout, stderr, _, err := cmd.Run(pt.ctx, workdir, nil, nil, nil, env, "version"); err != nil {
		ptLogF(t, "failed to get pulumi version: %v%s", err, formatCommandDiagnostics(stdout, stderr))
	} else {
		ptLogF(t, "pulumi version: %s", strings.TrimSpace(stdout))
	}
	if stdout, stderr, _, err := cmd.Run(pt.ctx, workdir, nil, nil, nil, env, "plugin", "ls", "-p"); err != nil {
		ptLogF(t, "failed to list pulumi plugins: %v%s", err, formatCommandDiagnostics(stdout, stderr))
	} else {
		ptLogF(t, "pulumi plugins:\n%s", strings.TrimSpace(stdout))
	}
}
