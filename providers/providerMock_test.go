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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestProviderMock(t *testing.T) {
	t.Parallel()
	source := filepath.Join("..", "pulumitest", "testdata", "python_gcp")

	t.Run("defaults", func(t *testing.T) {
		test := pulumitest.NewPulumiTest(t, source,
			opttest.AttachProvider("gcp",
				providers.ProviderMockFactory(providers.ProviderMocks{})))
		test.Preview()
	})

	t.Run("with mocks", func(t *testing.T) {
		var attached, configured, checkedConfig, checked, created bool
		test := pulumitest.NewPulumiTest(t, source,
			opttest.AttachProvider("gcp",
				providers.ProviderMockFactory(providers.ProviderMocks{
					Attach: func(ctx context.Context, in *pulumirpc.PluginAttach) (*emptypb.Empty, error) {
						attached = true
						return &emptypb.Empty{}, nil
					},
					Configure: func(ctx context.Context, in *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
						configured = true
						return &pulumirpc.ConfigureResponse{}, nil
					},
					CheckConfig: func(ctx context.Context, in *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
						checkedConfig = true
						return &pulumirpc.CheckResponse{}, nil
					},
					Check: func(ctx context.Context, in *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
						checked = true
						return &pulumirpc.CheckResponse{Inputs: in.News}, nil
					},
					Create: func(ctx context.Context, in *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
						created = true
						return &pulumirpc.CreateResponse{Id: "fake-id", Properties: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"url": {
									Kind: &structpb.Value_StringValue{StringValue: "fake-url"},
								},
							},
						}}, nil
					},
					Delete: func(ctx context.Context, in *pulumirpc.DeleteRequest) (*emptypb.Empty, error) {
						return &emptypb.Empty{}, nil
					},
				})))
		test.Preview()
		test.Up()
		assert.True(t, attached, "expected Attach to be called in mock")
		assert.True(t, configured, "expected Configure to be called in mock")
		assert.True(t, checkedConfig, "expected CheckConfig to be called in mock")
		assert.True(t, checked, "expected Check to be called in mock")
		assert.True(t, created, "expected Create to be called in mock")
	})
}
