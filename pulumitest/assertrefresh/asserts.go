package assertrefresh

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/changesummary"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// HasNoChanges asserts that the given RefreshResult has no changes.
func HasNoChanges(t *testing.T, up auto.RefreshResult) {
	t.Helper()

	summary := changesummary.FromStringIntMap(*up.Summary.ResourceChanges)
	unexpectedOps := summary.WhereOpNotEquals(apitype.OpSame)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s\n%s", unexpectedOps, up.StdOut)
	}
}
