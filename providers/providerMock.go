package providers

import (
	"context"
	"fmt"

	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/protobuf/types/known/emptypb"
)

type providerMock struct {
	rpc.UnimplementedResourceProviderServer

	mocks ProviderMocks
}

type ProviderMocks struct {
	Attach        func(ctx context.Context, in *rpc.PluginAttach) (*emptypb.Empty, error)
	Call          func(ctx context.Context, in *rpc.CallRequest) (*rpc.CallResponse, error)
	Cancel        func(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error)
	Check         func(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error)
	CheckConfig   func(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error)
	Configure     func(ctx context.Context, in *rpc.ConfigureRequest) (*rpc.ConfigureResponse, error)
	Construct     func(ctx context.Context, in *rpc.ConstructRequest) (*rpc.ConstructResponse, error)
	Create        func(ctx context.Context, in *rpc.CreateRequest) (*rpc.CreateResponse, error)
	Delete        func(ctx context.Context, in *rpc.DeleteRequest) (*emptypb.Empty, error)
	Diff          func(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error)
	DiffConfig    func(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error)
	GetMapping    func(ctx context.Context, in *rpc.GetMappingRequest) (*rpc.GetMappingResponse, error)
	GetMappings   func(ctx context.Context, in *rpc.GetMappingsRequest) (*rpc.GetMappingsResponse, error)
	GetPluginInfo func(ctx context.Context, in *emptypb.Empty) (*rpc.PluginInfo, error)
	GetSchema     func(ctx context.Context, in *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error)
	Invoke        func(ctx context.Context, in *rpc.InvokeRequest) (*rpc.InvokeResponse, error)
	Read          func(ctx context.Context, in *rpc.ReadRequest) (*rpc.ReadResponse, error)
	Update        func(ctx context.Context, in *rpc.UpdateRequest) (*rpc.UpdateResponse, error)
}

// ProviderInterceptFactory creates a new provider factory that can be used to intercept calls to a downstream provider.
func ProviderMockFactory(ctx context.Context, mocks ProviderMocks) ProviderFactory {
	return ResourceProviderFactory(func() (rpc.ResourceProviderServer, error) {
		return NewProviderMock(ctx, mocks)
	})
}

// NewProviderMock creates a new provider proxy that can be used to intercept calls to a downstream provider.
func NewProviderMock(ctx context.Context, mocks ProviderMocks) (rpc.ResourceProviderServer, error) {
	return &providerMock{
		mocks: mocks,
	}, nil
}

func (i *providerMock) Attach(ctx context.Context, in *rpc.PluginAttach) (*emptypb.Empty, error) {
	if i.mocks.Attach != nil {
		return i.mocks.Attach(ctx, in)
	}
	return &emptypb.Empty{}, nil
}

func (i *providerMock) Call(ctx context.Context, in *rpc.CallRequest) (*rpc.CallResponse, error) {
	if i.mocks.Call != nil {
		return i.mocks.Call(ctx, in)
	}
	return nil, fmt.Errorf("Call not mocked")
}

func (i *providerMock) Cancel(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	if i.mocks.Cancel != nil {
		return i.mocks.Cancel(ctx, in)
	}
	return &emptypb.Empty{}, nil
}

func (i *providerMock) Check(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	if i.mocks.Check != nil {
		return i.mocks.Check(ctx, in)
	}
	return &rpc.CheckResponse{
		Inputs: in.News,
	}, nil
}

func (i *providerMock) CheckConfig(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	if i.mocks.CheckConfig != nil {
		return i.mocks.CheckConfig(ctx, in)
	}
	return &rpc.CheckResponse{
		Inputs: in.News,
	}, nil
}

func (i *providerMock) Configure(ctx context.Context, in *rpc.ConfigureRequest) (*rpc.ConfigureResponse, error) {
	if i.mocks.Configure != nil {
		return i.mocks.Configure(ctx, in)
	}
	return &rpc.ConfigureResponse{
		AcceptSecrets:   true,
		SupportsPreview: true,
		AcceptResources: true,
		AcceptOutputs:   true,
	}, nil
}

func (i *providerMock) Construct(ctx context.Context, in *rpc.ConstructRequest) (*rpc.ConstructResponse, error) {
	if i.mocks.Construct != nil {
		return i.mocks.Construct(ctx, in)
	}
	return nil, fmt.Errorf("Construct not mocked")
}

func (i *providerMock) Create(ctx context.Context, in *rpc.CreateRequest) (*rpc.CreateResponse, error) {
	if i.mocks.Create != nil {
		return i.mocks.Create(ctx, in)
	}
	return nil, fmt.Errorf("Create not mocked")
}

func (i *providerMock) Delete(ctx context.Context, in *rpc.DeleteRequest) (*emptypb.Empty, error) {
	if i.mocks.Delete != nil {
		return i.mocks.Delete(ctx, in)
	}
	return nil, fmt.Errorf("Delete not mocked")
}

func (i *providerMock) Diff(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	if i.mocks.Diff != nil {
		return i.mocks.Diff(ctx, in)
	}
	return nil, fmt.Errorf("Diff not mocked")
}

func (i *providerMock) DiffConfig(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	if i.mocks.DiffConfig != nil {
		return i.mocks.DiffConfig(ctx, in)
	}
	return nil, fmt.Errorf("DiffConfig not mocked")
}

func (i *providerMock) GetMapping(ctx context.Context, in *rpc.GetMappingRequest) (*rpc.GetMappingResponse, error) {
	if i.mocks.GetMapping != nil {
		return i.mocks.GetMapping(ctx, in)
	}
	return nil, fmt.Errorf("GetMapping not mocked")
}

func (i *providerMock) GetMappings(ctx context.Context, in *rpc.GetMappingsRequest) (*rpc.GetMappingsResponse, error) {
	if i.mocks.GetMappings != nil {
		return i.mocks.GetMappings(ctx, in)
	}
	return nil, fmt.Errorf("GetMappings not mocked")
}

func (i *providerMock) GetPluginInfo(ctx context.Context, in *emptypb.Empty) (*rpc.PluginInfo, error) {
	if i.mocks.GetPluginInfo != nil {
		return i.mocks.GetPluginInfo(ctx, in)
	}
	return &rpc.PluginInfo{
		Version: "0.0.0",
	}, nil
}

func (i *providerMock) GetSchema(ctx context.Context, in *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {
	if i.mocks.GetSchema != nil {
		return i.mocks.GetSchema(ctx, in)
	}
	return &rpc.GetSchemaResponse{
		Schema: "{}",
	}, nil
}

func (i *providerMock) Invoke(ctx context.Context, in *rpc.InvokeRequest) (*rpc.InvokeResponse, error) {
	if i.mocks.Invoke != nil {
		return i.mocks.Invoke(ctx, in)
	}
	return nil, fmt.Errorf("Invoke not mocked")
}

func (i *providerMock) Read(ctx context.Context, in *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	if i.mocks.Read != nil {
		return i.mocks.Read(ctx, in)
	}
	return nil, fmt.Errorf("Read not mocked")
}

func (i *providerMock) Update(ctx context.Context, in *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	if i.mocks.Update != nil {
		return i.mocks.Update(ctx, in)
	}
	return nil, fmt.Errorf("Update not mocked")
}
