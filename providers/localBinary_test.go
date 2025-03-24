package providers_test

import (
	"context"
	"testing"

	"github.com/pulumi/pulumitest/providers"
	"github.com/stretchr/testify/assert"
)

type mockPulumiTest struct {
	source string
}

func (m *mockPulumiTest) Source() string {
	return m.source
}

func TestLocalBinaryAttach(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pt := &mockPulumiTest{source: t.TempDir()}
	factory := providers.DownloadPluginBinaryFactory("azure-native", "2.25.0")
	port, err := factory(ctx, pt)
	assert.NoError(t, err)
	assert.NotZero(t, port)
}
