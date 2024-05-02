package suite

import (
	"context"
	"fmt"
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

const (
	serverHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg          *config.Config
	LinkerClient linkerV1.LinkerClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/test.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cc, err := grpc.DialContext(ctx, serverAddress(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		Cfg:          cfg,
		LinkerClient: linkerV1.NewLinkerClient(cc),
	}
}

func serverAddress(cfg *config.Config) string {
	return net.JoinHostPort(serverHost, fmt.Sprint(cfg.Server.Port))
}
