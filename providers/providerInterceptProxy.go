package providers

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type providerInterceptProxy struct {
	rpc.UnimplementedResourceProviderServer

	client       rpc.ResourceProviderClient
	interceptors ProviderInterceptors
}

type ProviderInterceptors struct {
	Attach        func(ctx context.Context, in *rpc.PluginAttach, client rpc.ResourceProviderClient) (*emptypb.Empty, error)
	Call          func(ctx context.Context, in *rpc.CallRequest, client rpc.ResourceProviderClient) (*rpc.CallResponse, error)
	Cancel        func(ctx context.Context, in *emptypb.Empty, client rpc.ResourceProviderClient) (*emptypb.Empty, error)
	Check         func(ctx context.Context, in *rpc.CheckRequest, client rpc.ResourceProviderClient) (*rpc.CheckResponse, error)
	CheckConfig   func(ctx context.Context, in *rpc.CheckRequest, client rpc.ResourceProviderClient) (*rpc.CheckResponse, error)
	Configure     func(ctx context.Context, in *rpc.ConfigureRequest, client rpc.ResourceProviderClient) (*rpc.ConfigureResponse, error)
	Construct     func(ctx context.Context, in *rpc.ConstructRequest, client rpc.ResourceProviderClient) (*rpc.ConstructResponse, error)
	Create        func(ctx context.Context, in *rpc.CreateRequest, client rpc.ResourceProviderClient) (*rpc.CreateResponse, error)
	Delete        func(ctx context.Context, in *rpc.DeleteRequest, client rpc.ResourceProviderClient) (*emptypb.Empty, error)
	Diff          func(ctx context.Context, in *rpc.DiffRequest, client rpc.ResourceProviderClient) (*rpc.DiffResponse, error)
	DiffConfig    func(ctx context.Context, in *rpc.DiffRequest, client rpc.ResourceProviderClient) (*rpc.DiffResponse, error)
	GetMapping    func(ctx context.Context, in *rpc.GetMappingRequest, client rpc.ResourceProviderClient) (*rpc.GetMappingResponse, error)
	GetMappings   func(ctx context.Context, in *rpc.GetMappingsRequest, client rpc.ResourceProviderClient) (*rpc.GetMappingsResponse, error)
	GetPluginInfo func(ctx context.Context, in *emptypb.Empty, client rpc.ResourceProviderClient) (*rpc.PluginInfo, error)
	GetSchema     func(ctx context.Context, in *rpc.GetSchemaRequest, client rpc.ResourceProviderClient) (*rpc.GetSchemaResponse, error)
	Invoke        func(ctx context.Context, in *rpc.InvokeRequest, client rpc.ResourceProviderClient) (*rpc.InvokeResponse, error)
	Read          func(ctx context.Context, in *rpc.ReadRequest, client rpc.ResourceProviderClient) (*rpc.ReadResponse, error)
	Update        func(ctx context.Context, in *rpc.UpdateRequest, client rpc.ResourceProviderClient) (*rpc.UpdateResponse, error)
}

// ProviderInterceptFactory creates a new provider factory that can be used to intercept calls to a downstream provider.
func ProviderInterceptFactory(ctx context.Context, factory ProviderFactory, interceptors ProviderInterceptors) ProviderFactory {
	return ResourceProviderFactory(func() (rpc.ResourceProviderServer, error) {
		port, err := factory(ctx)
		if err != nil {
			return nil, err
		}
		return NewProviderInterceptProxy(ctx, port, interceptors)
	})
}

// NewProviderInterceptProxy creates a new provider proxy that can be used to intercept calls to a downstream provider.
func NewProviderInterceptProxy(ctx context.Context, downstreamProviderPort Port, interceptors ProviderInterceptors) (rpc.ResourceProviderServer, error) {
	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("127.0.0.1:%d", downstreamProviderPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(rpcutil.OpenTracingClientInterceptor()),
		grpc.WithStreamInterceptor(rpcutil.OpenTracingStreamClientInterceptor()))
	if err != nil {
		return nil, err
	}
	client := rpc.NewResourceProviderClient(conn)
	return &providerInterceptProxy{
		client:       client,
		interceptors: interceptors,
	}, nil
}

func (i *providerInterceptProxy) Attach(ctx context.Context, in *rpc.PluginAttach) (*emptypb.Empty, error) {
	if i.interceptors.Attach != nil {
		return i.interceptors.Attach(ctx, in, i.client)
	}
	return i.client.Attach(ctx, in)
}

func (i *providerInterceptProxy) Call(ctx context.Context, in *rpc.CallRequest) (*rpc.CallResponse, error) {
	if i.interceptors.Call != nil {
		return i.interceptors.Call(ctx, in, i.client)
	}
	return i.client.Call(ctx, in)
}

func (i *providerInterceptProxy) Cancel(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	if i.interceptors.Cancel != nil {
		return i.interceptors.Cancel(ctx, in, i.client)
	}
	return i.client.Cancel(ctx, in)
}

func (i *providerInterceptProxy) Check(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	if i.interceptors.Check != nil {
		return i.interceptors.Check(ctx, in, i.client)
	}
	return i.client.Check(ctx, in)
}

func (i *providerInterceptProxy) CheckConfig(ctx context.Context, in *rpc.CheckRequest) (*rpc.CheckResponse, error) {
	if i.interceptors.CheckConfig != nil {
		return i.interceptors.CheckConfig(ctx, in, i.client)
	}
	return i.client.CheckConfig(ctx, in)
}

func (i *providerInterceptProxy) Configure(ctx context.Context, in *rpc.ConfigureRequest) (*rpc.ConfigureResponse, error) {
	if i.interceptors.Configure != nil {
		return i.interceptors.Configure(ctx, in, i.client)
	}
	return i.client.Configure(ctx, in)
}

func (i *providerInterceptProxy) Construct(ctx context.Context, in *rpc.ConstructRequest) (*rpc.ConstructResponse, error) {
	if i.interceptors.Construct != nil {
		return i.interceptors.Construct(ctx, in, i.client)
	}
	return i.client.Construct(ctx, in)
}

func (i *providerInterceptProxy) Create(ctx context.Context, in *rpc.CreateRequest) (*rpc.CreateResponse, error) {
	if i.interceptors.Create != nil {
		return i.interceptors.Create(ctx, in, i.client)
	}
	return i.client.Create(ctx, in)
}

func (i *providerInterceptProxy) Delete(ctx context.Context, in *rpc.DeleteRequest) (*emptypb.Empty, error) {
	if i.interceptors.Delete != nil {
		return i.interceptors.Delete(ctx, in, i.client)
	}
	return i.client.Delete(ctx, in)
}

func (i *providerInterceptProxy) Diff(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	if i.interceptors.Diff != nil {
		return i.interceptors.Diff(ctx, in, i.client)
	}
	return i.client.Diff(ctx, in)
}

func (i *providerInterceptProxy) DiffConfig(ctx context.Context, in *rpc.DiffRequest) (*rpc.DiffResponse, error) {
	if i.interceptors.DiffConfig != nil {
		return i.interceptors.DiffConfig(ctx, in, i.client)
	}
	return i.client.DiffConfig(ctx, in)
}

func (i *providerInterceptProxy) GetMapping(ctx context.Context, in *rpc.GetMappingRequest) (*rpc.GetMappingResponse, error) {
	if i.interceptors.GetMapping != nil {
		return i.interceptors.GetMapping(ctx, in, i.client)
	}
	return i.client.GetMapping(ctx, in)
}

func (i *providerInterceptProxy) GetMappings(ctx context.Context, in *rpc.GetMappingsRequest) (*rpc.GetMappingsResponse, error) {
	if i.interceptors.GetMappings != nil {
		return i.interceptors.GetMappings(ctx, in, i.client)
	}
	return i.client.GetMappings(ctx, in)
}

func (i *providerInterceptProxy) GetPluginInfo(ctx context.Context, in *emptypb.Empty) (*rpc.PluginInfo, error) {
	if i.interceptors.GetPluginInfo != nil {
		return i.interceptors.GetPluginInfo(ctx, in, i.client)
	}
	return i.client.GetPluginInfo(ctx, in)
}

func (i *providerInterceptProxy) GetSchema(ctx context.Context, in *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {
	if i.interceptors.GetSchema != nil {
		return i.interceptors.GetSchema(ctx, in, i.client)
	}
	return i.client.GetSchema(ctx, in)
}

func (i *providerInterceptProxy) Invoke(ctx context.Context, in *rpc.InvokeRequest) (*rpc.InvokeResponse, error) {
	if i.interceptors.Invoke != nil {
		return i.interceptors.Invoke(ctx, in, i.client)
	}
	return i.client.Invoke(ctx, in)
}

func (i *providerInterceptProxy) Read(ctx context.Context, in *rpc.ReadRequest) (*rpc.ReadResponse, error) {
	if i.interceptors.Read != nil {
		return i.interceptors.Read(ctx, in, i.client)
	}
	return i.client.Read(ctx, in)
}

func (i *providerInterceptProxy) Update(ctx context.Context, in *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	if i.interceptors.Update != nil {
		return i.interceptors.Update(ctx, in, i.client)
	}
	return i.client.Update(ctx, in)
}
