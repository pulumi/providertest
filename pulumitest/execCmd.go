package pulumitest

import (
	"bytes"
	"context"
	"fmt"
)

type cmdOutput struct {
	Args  []string
	Stdout string
	Stderr string
	ReturnCode int
}

func (a *PulumiTest) execCmd(args ...string) cmdOutput {
	a.t.Helper()
	workspace := a.CurrentStack().Workspace()
	ctx := context.Background()
	workdir := workspace.WorkDir()
	var env []string
	for k, v := range workspace.GetEnvVars() {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	stdin := bytes.NewReader([]byte{})

	s1, s2, code, err := workspace.PulumiCommand().Run(ctx, workdir, stdin, nil, nil, env, args...)
	if err != nil {
		a.logf(s1)
		a.logf(s2)
		a.fatalf("Failed to run command %v: %v", args, err)
	}

	return cmdOutput{
		Args: args,
		Stdout: s1,
		Stderr: s2,
		ReturnCode: code,
	}
}

