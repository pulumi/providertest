package autotest

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremove"
)

// NewStack creates a new stack, ensure it's cleaned up after the test is done.
// If no stack name is provided, a random one will be generated.
func (a *AutoTest) NewStack(stackName string, opts ...auto.LocalWorkspaceOption) *auto.Stack {
	a.t.Helper()

	if stackName == "" {
		stackName = randomStackName(a.source)
	}

	// Set default stack opts. These can be overridden by the caller.
	stackOpts := []auto.LocalWorkspaceOption{
		auto.EnvVars(a.Env().GetEnv()),
	}
	stackOpts = append(stackOpts, opts...)

	stack, err := auto.NewStackLocalSource(a.ctx, stackName, a.source, stackOpts...)

	if err != nil {
		a.t.Fatalf("failed to create stack: %s", err)
		return nil
	}
	a.t.Cleanup(func() {
		a.t.Helper()
		a.t.Log("cleaning up stack")
		_, err := stack.Destroy(a.ctx)
		if err != nil {
			a.t.Errorf("failed to destroy stack: %s", err)
		}
		err = stack.Workspace().RemoveStack(a.ctx, stackName, optremove.Force())
		if err != nil {
			a.t.Errorf("failed to remove stack: %s", err)
		}
	})
	a.currentStack = &stack
	return &stack
}

func randomStackName(dir string) string {
	// Fetch the host and test dir names, cleaned so to contain just [a-zA-Z0-9-_] chars.
	hostname, err := os.Hostname()
	contract.AssertNoErrorf(err, "failure to fetch hostname for stack prefix")
	var host string
	for _, c := range hostname {
		if len(host) >= 10 {
			break
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			host += string(c)
		}
	}

	var test string
	for _, c := range filepath.Base(dir) {
		if len(test) >= 10 {
			break
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			test += string(c)
		}
	}

	b := make([]byte, 4)
	_, err = rand.Read(b)
	contract.AssertNoErrorf(err, "failure to generate random stack suffix")

	return strings.ToLower("p-it-" + host + "-" + test + "-" + hex.EncodeToString(b))

}
