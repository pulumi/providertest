package assertpreview

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/changesummary"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func HasNoChanges(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := changesummary.ChangeSummary(preview.ChangeSummary)
	unexpectedOps := convertedMap.WhereOpNotEquals(apitype.OpSame)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s\n%s", unexpectedOps, preview.StdOut)
	}
}

func HasNoDeletes(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := changesummary.ChangeSummary(preview.ChangeSummary)
	unexpectedOps := convertedMap.WhereOpEquals(apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s\n%s", unexpectedOps, preview.StdOut)
	}
}

func HasNoReplacements(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := changesummary.ChangeSummary(preview.ChangeSummary)
	unexpectedOps := convertedMap.WhereOpEquals(apitype.OpReplace, apitype.OpCreateReplacement, apitype.OpDeleteReplaced, apitype.OpDiscardReplaced, apitype.OpImportReplacement, apitype.OpReadReplacement)

	if len(*unexpectedOps) > 0 {
		t.Errorf("expected no replacements, got %s\n%s", unexpectedOps, preview.StdOut)
	}
}
