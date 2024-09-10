package grpcapp

import (
	"fmt"
	urlShortener "github.com/Sleeps17/linker/internal/clients/url-shortener"
	server "github.com/Sleeps17/linker/internal/grpc/linker"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log    *slog.Logger
	server *grpc.Server
	port   int
}

func New(
	log *slog.Logger,
	port int,
	linkerService server.LinkService,
	topicService server.TopicService,
	urlShortener urlShortener.UrlShortener,
) *App {
	grpcServer := grpc.NewServer()

	server.Register(grpcServer, log, linkerService, topicService, urlShortener)

	return &App{
		log:    log,
		server: grpcServer,
		port:   port,
	}
}

func (a *App) MustRun() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		panic(fmt.Sprintf("Failed to listen: %v", err))
	}

	if err := a.server.Serve(l); err != nil {
		panic(fmt.Sprintf("Failed to serve: %v", err))
	}
}

func (a *App) Stop() {
	a.server.GracefulStop()
}
