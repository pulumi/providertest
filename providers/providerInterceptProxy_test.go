package providers_test

import (
	"context"
	"path/filepath"
	"testing"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/pulumi/pulumitest"
	"github.com/pulumi/pulumitest/opttest"
	"github.com/pulumi/pulumitest/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProviderInterceptProxy(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	var didAttach bool
	// Ensure plugin is downloaded so YAML can look up its schema
	_, err := providers.DownloadPluginBinary("azure-native", "2.10.0")
	require.NoError(t, err)

	interceptedFactory := providers.ProviderInterceptFactory(ctx, providers.DownloadPluginBinaryFactory("azure-native", "2.10.0"), providers.ProviderInterceptors{
		Attach: func(ctx context.Context, in *pulumirpc.PluginAttach, client pulumirpc.ResourceProviderClient) (*emptypb.Empty, error) {
			didAttach = true
			return client.Attach(ctx, in)
		},
		Configure: func(ctx context.Context, in *pulumirpc.ConfigureRequest, client pulumirpc.ResourceProviderClient) (*pulumirpc.ConfigureResponse, error) {
			// Skip checking the real configuration
			return &pulumirpc.ConfigureResponse{}, nil
		},
		Check: func(ctx context.Context, in *pulumirpc.CheckRequest, client pulumirpc.ResourceProviderClient) (*pulumirpc.CheckResponse, error) {
			// Skip checking the real configuration
			return &pulumirpc.CheckResponse{Inputs: in.News}, nil
		},
	})
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("..", "testdata", "yaml_azure"),
		opttest.AttachProvider("azure-native", interceptedFactory))

	test.Preview(t)
	assert.True(t, didAttach, "expected Attach to be called in proxy")
}
