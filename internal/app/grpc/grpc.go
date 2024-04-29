package grpcApp

import (
	"fmt"
	server "github.com/Sleeps17/linker/internal/grpc/linker"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log    *slog.Logger
	server *grpc.Server
	host   string
	port   int
}

func New(log *slog.Logger, host string, port int, linkerService server.LinkerService) *App {
	grpcServer := grpc.NewServer()

	server.Register(grpcServer, linkerService)

	return &App{
		log:    log,
		server: grpcServer,
		host:   host,
		port:   port,
	}
}

func (app *App) MustRun() {
	l, err := net.Listen("tcp", net.JoinHostPort(app.host, fmt.Sprint(app.port)))
	if err != nil {
		panic(fmt.Sprintf("Failed to listen: %v", err))
	}

	if err := app.server.Serve(l); err != nil {
		panic(fmt.Sprintf("Failed to serve: %v", err))
	}
}

func (a *App) Stop() {
	a.server.GracefulStop()
}
