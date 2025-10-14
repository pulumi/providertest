package pulumitest

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest/assertup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

// TestDotNetAwsDeploy tests deploying a real AWS S3 bucket using C# and AWS SSO credentials
func TestDotNetAwsDeploy(t *testing.T) {
	// Skip in CI or if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()
	test := NewPulumiTest(t, filepath.Join("testdata", "csharp_aws"))

	// Test a preview - should show 2 resources (stack + bucket)
	preview := test.Preview(t)
	assert.Equal(t,
		map[apitype.OpType]int{apitype.OpCreate: 2},
		preview.ChangeSummary)
	t.Logf("Preview showed %d resources to create", len(preview.ChangeSummary))

	// Now do a real deploy
	up := test.Up(t)
	assert.Equal(t,
		map[string]int{"create": 2},
		*up.Summary.ResourceChanges)

	assertup.HasNoDeletes(t, up)

	// Verify outputs exist
	assert.NotEmpty(t, up.Outputs["bucketName"].Value)
	assert.NotEmpty(t, up.Outputs["bucketArn"].Value)
	t.Logf("Created S3 bucket with ARN: %v", up.Outputs["bucketArn"].Value)

	// Test that a second preview shows no changes (resources will show as "same")
	preview2 := test.Preview(t)
	// Should show 2 resources as "same" (stack + bucket) with no creates/updates/deletes
	assert.NotContains(t, preview2.ChangeSummary, apitype.OpCreate)
	assert.NotContains(t, preview2.ChangeSummary, apitype.OpUpdate)
	assert.NotContains(t, preview2.ChangeSummary, apitype.OpDelete)
	t.Logf("Second preview confirmed no changes needed - %d resources unchanged", preview2.ChangeSummary[apitype.OpSame])

	// The cleanup will automatically destroy the bucket via t.Cleanup()
	t.Log("Test completed successfully - bucket will be destroyed automatically")
}
