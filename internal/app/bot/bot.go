package botapp

import (
	linkerbot "github.com/Sleeps17/linker/internal/bot/linker"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/storage"
	"log/slog"
)

type App struct {
	bot *linkerbot.Bot
	log *slog.Logger
}

func MustNew(cfg *config.BotConfig, log *slog.Logger, storage storage.Storage) *App {
	bot, err := linkerbot.New(cfg, log, storage, storage)
	if err != nil {
		panic(err)
	}

	return &App{
		bot: bot,
		log: log,
	}
}

func (a *App) MustRun() {
	a.bot.MustRun()
}

func (a *App) Stop() {
	a.bot.Stop()
}
