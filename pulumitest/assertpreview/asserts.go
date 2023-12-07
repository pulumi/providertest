package assertpreview

import (
	"strings"
	"testing"

	"github.com/pulumi/providertest/pulumitest/asserthelpers"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func HasNoChanges(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := asserthelpers.OpMapToStringMap(preview.ChangeSummary)
	unexpectedOps := asserthelpers.FindUnexpectedOps(convertedMap, apitype.OpSame)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

func HasNoDeletes(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := asserthelpers.OpMapToStringMap(preview.ChangeSummary)
	unexpectedOps := asserthelpers.FindMatchingOps(convertedMap, apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}
