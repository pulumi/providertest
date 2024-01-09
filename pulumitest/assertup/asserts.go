package assertup

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/changesummary"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// UpHasNoChanges asserts that the given UpResult has no changes - only "same" operations allowed.
func HasNoChanges(t *testing.T, up auto.UpResult) {
	t.Helper()

	summary := changesummary.FromStringIntMap(*up.Summary.ResourceChanges)
	unexpectedOps := summary.WhereOpNotEquals(apitype.OpSame)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s\n%s", unexpectedOps, up.StdOut)
	}
}

// UpHasNoDeletes asserts that the given UpResult has no deletes - only "same", "create", "update", "refresh", and
func HasNoDeletes(t *testing.T, up auto.UpResult) {
	t.Helper()

	summary := changesummary.FromStringIntMap(*up.Summary.ResourceChanges)
	unexpectedOps := summary.WhereOpEquals(apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s\n%s", unexpectedOps, up.StdOut)
	}
}

func HasNoReplacements(t *testing.T, up auto.UpResult) {
	t.Helper()

	summary := changesummary.FromStringIntMap(*up.Summary.ResourceChanges)
	unexpectedOps := summary.WhereOpEquals(apitype.OpReplace, apitype.OpCreateReplacement, apitype.OpDeleteReplaced, apitype.OpDiscardReplaced, apitype.OpImportReplacement, apitype.OpReadReplacement)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no replacements, got %s\n%s", unexpectedOps, up.StdOut)
	}
}
