package pulumitest_test

import (
	"testing"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/stretchr/testify/assert"
)

func TestGrpc(t *testing.T) {
	t.Parallel()
	test := pulumitest.NewPulumiTest(t, "testdata/yaml_program")
	test.Up()
	log := test.GrpcLog()
	assert.NotEmpty(t, log)
	creates, err := log.Creates()
	assert.NoError(t, err)
	assert.Len(t, creates, 1)
	snaps.MatchJSON(t, creates, match.Any("0.Response.id", "0.Response.properties.id"))
}
