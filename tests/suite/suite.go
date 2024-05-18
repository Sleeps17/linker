package suite

import (
	"context"
	"errors"
	"fmt"
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/internal/app"
	mockUrlShortener "github.com/Sleeps17/linker/internal/clients/url-shortener/mock"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/logger"
	"github.com/Sleeps17/linker/internal/storage/postgresql"
	"github.com/golang/mock/gomock"
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

	cfg := config.MustLoadByPath("../config/test.yaml")

	log := logger.Setup(cfg.Env)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DataBase.Timeout)
	defer cancel()
	storage := postgresql.MustNew(ctx, createPostgresConnString(cfg))

	ctrl := gomock.NewController(t)
	mockedShortener := mockUrlShortener.NewMockUrlShortener(ctrl)
	mockedShortener.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("some err")).AnyTimes()
	mockedShortener.EXPECT().DeleteURL(gomock.Any(), gomock.Any()).Return(errors.New("some err")).AnyTimes()

	application := app.New(log, int(cfg.Server.Port), storage, storage, mockedShortener)
	log.Info("application configured successfully")

	// TODO: Start server
	go application.MustRun()

	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.Timeout)

	t.Cleanup(func() {
		t.Helper()
		ctrl.Finish()
		cancel()

		application.Stop()
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

func createPostgresConnString(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.DataBase.Host,
		cfg.DataBase.Port,
		cfg.DataBase.Username,
		cfg.DataBase.Name,
		cfg.DataBase.Password,
	)
}
