package app

import (
	grpcApp "github.com/Sleeps17/linker/internal/app/grpc"
	server "github.com/Sleeps17/linker/internal/grpc/linker"
	"log/slog"
)

type App struct {
	grpcSrv *grpcApp.App
}

func New(log *slog.Logger, host string, port int, linkerService server.LinkerService) *App {

	return &App{grpcSrv: grpcApp.New(log, host, port, linkerService)}
}

func (a *App) MustRun() {
	a.grpcSrv.MustRun()
}

func (a *App) Stop() {
	a.grpcSrv.Stop()
}
