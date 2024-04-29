package main

import (
	"context"
	"github.com/Sleeps17/linker/internal/app"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/logger"
	"github.com/Sleeps17/linker/internal/storage/mongodb"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// TODO: Load config
	cfg := config.MustLoad()

	// TODO: Init logger
	log := logger.Setup(cfg.Env)

	// TODO: Init DB
	storage := mongodb.MustNew(context.Background(), cfg.DataBase.ConnString, cfg.DataBase.DbName, cfg.DataBase.Collection)

	// TODO: Init server
	app := app.New(log, cfg.Server.Host, int(cfg.Server.Port), storage)

	// TODO: Start server
	go app.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	app.Stop()
}
