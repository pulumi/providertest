package providers

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
)

type ResourceProviderServerFactory func(PulumiTest) (pulumirpc.ResourceProviderServer, error)

// startProvider starts the provider in a goProc and returns the port it's listening on.
// To shut down the provider, cancel the context.
func ResourceProviderFactory(makeResourceProviderServer ResourceProviderServerFactory) ProviderFactory {
	return func(ctx context.Context, pt PulumiTest) (Port, error) {
		return startResourceProviderServer(ctx, pt, makeResourceProviderServer)
	}
}

func startResourceProviderServer(ctx context.Context, pt PulumiTest, makeResourceProviderServer ResourceProviderServerFactory) (Port, error) {
	cancelChannel := make(chan bool)
	go func() {
		<-ctx.Done()
		close(cancelChannel)
	}()

	handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
		Cancel: cancelChannel,
		Init: func(srv *grpc.Server) error {
			prov, proverr := makeResourceProviderServer(pt)
			if proverr != nil {
				return fmt.Errorf("failed to create resource provider server: %v", proverr)
			}
			pulumirpc.RegisterResourceProviderServer(srv, prov)
			return nil
		},
		Options: rpcutil.OpenTracingServerInterceptorOptions(nil),
	})
	if err != nil {
		return 0, fmt.Errorf("fatal: %v", err)
	}

	return Port(handle.Port), nil
}
