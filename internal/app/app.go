package app

import (
	grpcApp "github.com/Sleeps17/linker/internal/app/grpc"
	server "github.com/Sleeps17/linker/internal/grpc/linker"
	"log/slog"
)

type App struct {
	grpcSrv *grpcApp.App
}

func New(log *slog.Logger, port int, linkerService server.Service) *App {

	return &App{grpcSrv: grpcApp.New(log, port, linkerService)}
}

func (a *App) MustRun() {
	a.grpcSrv.MustRun()
}

func (a *App) Stop() {
	a.grpcSrv.Stop()
}
