package replay

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/providertest/providers"
)

// When testing methods that return CheckFailure lists, the tests do not want to care about the
// ordering of the individual failures. These tests should not or fail when providers change the
// ordering or exhibit non-determinism. To accomplish this, ensure that Replay sorts CheckFailures
// before matching against the expected response.
func TestReplayNormalizesCheckFailureOrder(t *testing.T) {
	failures := []*pulumirpc.CheckFailure{
		{Property: "B", Reason: "B-failed"},
		{Property: "A", Reason: "A-failed"},
	}

	p, err := providers.NewProviderMock(providers.ProviderMocks{
		CheckConfig: func(ctx context.Context, in *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
			return &pulumirpc.CheckResponse{Failures: failures}, nil
		},
		Check: func(ctx context.Context, in *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
			return &pulumirpc.CheckResponse{Failures: failures}, nil
		},
		Invoke: func(ctx context.Context, in *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
			return &pulumirpc.InvokeResponse{Failures: failures}, nil
		},
		Call: func(ctx context.Context, in *pulumirpc.CallRequest) (*pulumirpc.CallResponse, error) {
			return &pulumirpc.CallResponse{Failures: failures}, nil
		},
	})
	require.NoError(t, err)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Check",
	  "request": {
	    "urn": "u",
	    "news": {}
	  },
	  "response": {
	    "failures": [
	      {
		"property": "A",
		"reason": "A-failed"
	      },
	      {
		"property": "B",
		"reason": "B-failed"
	      }
	    ]
	  }
	}`)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/CheckConfig",
	  "request": {
	    "urn": "u",
	    "news": {}
	  },
	  "response": {
	    "failures": [
	      {
		"property": "A",
		"reason": "A-failed"
	      },
	      {
		"property": "B",
		"reason": "B-failed"
	      }
	    ]
	  }
	}`)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Invoke",
	  "request": {
	    "tok": "t",
	    "args": {}
	  },
	  "response": {
	    "failures": [
	      {
		"property": "A",
		"reason": "A-failed"
	      },
	      {
		"property": "B",
		"reason": "B-failed"
	      }
	    ]
	  }
	}`)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Call",
	  "request": {
	    "tok": "t",
	    "args": {}
	  },
	  "response": {
	    "failures": [
	      {
		"property": "A",
		"reason": "A-failed"
	      },
	      {
		"property": "B",
		"reason": "B-failed"
	      }
	    ]
	  }
	}`)
}

func TestMatchingErrors(t *testing.T) {
	p, err := providers.NewProviderMock(providers.ProviderMocks{
		Check: func(ctx context.Context, in *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
			return &pulumirpc.CheckResponse{}, fmt.Errorf("An error has occurred")
		},
	})
	require.NoError(t, err)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Check",
	  "request": {
	    "urn": "u",
	    "news": {}
	  },
          "errors": ["An error has occurred"]
	}`)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Check",
	  "request": {
	    "urn": "u",
	    "news": {}
	  },
          "errors": ["An error has occurred", "Unrelated error"]
	}`)

	Replay(t, p, `
	{
	  "method": "/pulumirpc.ResourceProvider/Check",
	  "request": {
	    "urn": "u",
	    "news": {}
	  },
          "errors": ["*"]
	}`)
}

func TestHandshake(t *testing.T) {
	p, err := providers.NewProviderMock(providers.ProviderMocks{
		Handshake: func(ctx context.Context, in *pulumirpc.ProviderHandshakeRequest) (*pulumirpc.ProviderHandshakeResponse, error) {
			return &pulumirpc.ProviderHandshakeResponse{}, nil
		},
	})

	require.NoError(t, err)

	Replay(t, p, `
	  {
    "method": "/pulumirpc.ResourceProvider/Handshake",
    "request": {
      "engineAddress": "127.0.0.1:57575",
      "rootDirectory": "/Users/vvm/code/pulumi-terraform-bridge/pkg/pf/tests/testdatagen/genrandom/bin",
      "programDirectory": "/Users/vvm/code/pulumi-terraform-bridge/pkg/pf/tests/testdatagen/genrandom/bin"
    },
    "response": {
    },
    "metadata": {
      "kind": "resource",
      "mode": "client",
      "name": "random"
    }
  }`)
}

func TestHandshakeError(t *testing.T) {
	p, err := providers.NewProviderMock(providers.ProviderMocks{
		Handshake: func(ctx context.Context, in *pulumirpc.ProviderHandshakeRequest) (*pulumirpc.ProviderHandshakeResponse, error) {
			return &pulumirpc.ProviderHandshakeResponse{}, fmt.Errorf("An error has occurred")
		},
	})

	require.NoError(t, err)

	Replay(t, p, `
	  {
    "method": "/pulumirpc.ResourceProvider/Handshake",
    "request": {
      "engineAddress": "127.0.0.1:57575",
      "rootDirectory": "/Users/vvm/code/pulumi-terraform-bridge/pkg/pf/tests/testdatagen/genrandom/bin",
      "programDirectory": "/Users/vvm/code/pulumi-terraform-bridge/pkg/pf/tests/testdatagen/genrandom/bin"
    },
    "errors": ["An error has occurred"],
    "metadata": {
      "kind": "resource",
      "mode": "client",
      "name": "random"
    }
  }`)
}
