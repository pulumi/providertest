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
	var didAttach, didForward bool
	mockFactory := providers.ProviderMockFactory(ctx, providers.ProviderMocks{
		Attach: func(ctx context.Context, in *pulumirpc.PluginAttach) (*emptypb.Empty, error) {
			didForward = true
			return &emptypb.Empty{}, nil
		},
		Create: func(ctx context.Context, in *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
			// Return without error to indicate success.
			return &pulumirpc.CreateResponse{}, nil
		},
	})
	interceptedFactory := providers.ProviderInterceptFactory(ctx, mockFactory, providers.ProviderInterceptors{
		Attach: func(ctx context.Context, in *pulumirpc.PluginAttach, client pulumirpc.ResourceProviderClient) (*emptypb.Empty, error) {
			didAttach = true
			return client.Attach(ctx, in)
		},
	})
	test := pulumitest.NewPulumiTest(t, filepath.Join("..", "pulumitest", "testdata", "yaml_program"), opttest.AttachProvider("random", interceptedFactory))

	test.Preview()
	assert.True(t, didAttach, "expected Attach to be called in proxy")
	assert.True(t, didForward, "expected Attach to be called in downstream provider")
}
