package providers_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProviderInterceptProxy(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	didAttach := false
	gcpProviderFactory := providers.DownloadPluginBinaryFactory("azure-native", "2.21.0")
	interceptedFactory := providers.ProviderInterceptFactory(ctx, gcpProviderFactory, providers.ProviderInterceptors{
		Attach: func(ctx context.Context, in *pulumirpc.PluginAttach, client pulumirpc.ResourceProviderClient) (*emptypb.Empty, error) {
			didAttach = true
			return client.Attach(ctx, in)
		},
	})
	test := pulumitest.NewPulumiTest(t, filepath.Join("..", "pulumitest", "testdata", "yaml_azure"), opttest.AttachProvider("azure-native", interceptedFactory))
	test.SetConfig("azure-native:location", "WestUS2")
	test.Preview()
	assert.True(t, didAttach, "expected Attach to be called")
}
