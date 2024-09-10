package httpapp

import (
	"context"
	"errors"
	"github.com/Sleeps17/linker/internal/config"
	httpserver "github.com/Sleeps17/linker/internal/http/linker"
	handlers2 "github.com/Sleeps17/linker/internal/http/linker/handlers"
	"github.com/Sleeps17/linker/internal/storage"
	"log/slog"
	"net/http"
)

type App struct {
	log *slog.Logger
	srv *httpserver.Server
	cfg *config.ServerConfig
}

func New(cfg *config.ServerConfig, log *slog.Logger, storage storage.Storage) *App {
	topicHandler := handlers2.NewTopicHandler(log, storage)
	linkHandler := handlers2.NewLinkHandler(log, storage)

	srv := httpserver.NewServer(cfg, topicHandler, linkHandler)

	return &App{
		log: log,
		srv: srv,
		cfg: cfg,
	}
}

func (a *App) MustRun() {
	a.log.Info("http server started", slog.String("address", a.cfg.Port))

	if err := a.srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (a *App) Stop() {
	a.log.Info("http server stopped")
	_ = a.srv.Stop(context.Background())
}
