package main

import (
	"context"
	"fmt"
	"github.com/Sleeps17/linker/internal/app"
	urlShortenerClient "github.com/Sleeps17/linker/internal/clients/url-shortener/url-shortener-client"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/logger"
	"github.com/Sleeps17/linker/internal/storage/postgresql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// TODO: Load config
	cfg := config.MustLoad()

	// TODO: Init logger
	log := logger.Setup(cfg.Env)
	log.Info("logger configured successfully", slog.String("env", cfg.Env))

	// TODO: Init DB
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DataBase.Timeout)
	defer cancel()
	storage := postgresql.MustNew(ctx, createPostgresConnString(cfg))
	log.Info("database configured successfully", slog.String("db_name", cfg.DataBase.Name))

	urlShortener := urlShortenerClient.New(
		cfg.UrlShortenerClient.Host,
		cfg.UrlShortenerClient.Port,
		cfg.UrlShortenerClient.Username,
		cfg.UrlShortenerClient.Password,
	)

	// TODO: Init server
	application := app.New(log, int(cfg.Server.Port), storage, storage, urlShortener)
	log.Info("application configured successfully")

	// TODO: Start server
	go application.MustRun()
	log.Info("application started successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	application.Stop()
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
