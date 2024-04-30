package main

import (
	"context"
	"github.com/Sleeps17/linker/internal/app"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/logger"
	"github.com/Sleeps17/linker/internal/storage/mongodb"
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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DataBase.ConnectionTimeout)
	defer cancel()
	storage := mongodb.MustNew(ctx, cfg.DataBase.ConnString, cfg.DataBase.DbName, cfg.DataBase.Collection)
	log.Info("database configured successfully", slog.String("db_name", cfg.DataBase.DbName))

	// TODO: Init server
	application := app.New(log, int(cfg.Server.Port), storage)
	log.Info("application configured successfully")

	// TODO: Start server
	go application.MustRun()
	log.Info("application started successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	application.Stop()
}
