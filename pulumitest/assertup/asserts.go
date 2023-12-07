package assertup

import (
	"strings"
	"testing"

	"github.com/pulumi/providertest/pulumitest/asserthelpers"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// UpHasNoChanges asserts that the given UpResult has no changes - only "same" operations allowed.
func HasNoChanges(t *testing.T, up auto.UpResult) {
	t.Helper()

	unexpectedOps := asserthelpers.FindUnexpectedOps(*up.Summary.ResourceChanges, apitype.OpSame)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

// UpHasNoDeletes asserts that the given UpResult has no deletes - only "same", "create", "update", "refresh", and
func HasNoDeletes(t *testing.T, up auto.UpResult) {
	t.Helper()

	unexpectedOps := asserthelpers.FindMatchingOps(*up.Summary.ResourceChanges, apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}
