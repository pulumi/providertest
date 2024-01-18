package providers_test

import (
	"context"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/stretchr/testify/assert"
)

func TestLocalBinaryAttach(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	factory := providers.DownloadPluginBinaryFactory("azure-native", "2.25.0")
	port, err := factory(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, port)
}
