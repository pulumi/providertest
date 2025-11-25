package pulumitest_test

import (
	"testing"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/stretchr/testify/assert"
)

func TestGrpc(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")

	versionCmd := test.CurrentStack().Workspace().PulumiCommand()
	version, _, _, err := versionCmd.Run(test.Context(), test.WorkingDir(), nil, nil, nil, nil, "version")
	contract.AssertNoErrorf(err, "failed to get pulumi version: %s", version)
	t.Logf("pulumi version: %s", version)

	test.Up(t)
	log := test.GrpcLog(t)
	assert.NotEmpty(t, log)
	creates, err := log.Creates()
	assert.NoError(t, err)
	assert.Len(t, creates, 1)
	snaps.MatchJSON(t, creates, match.Any("0.Response.id", "0.Response.properties.id"))
}
