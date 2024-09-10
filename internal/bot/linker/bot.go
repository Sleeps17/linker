package linkerbot

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	bothandlers "github.com/Sleeps17/linker/internal/bot/linker/handlers"
	"github.com/Sleeps17/linker/internal/config"
	"log/slog"
)

type Bot struct {
	api      *gotgbot.Bot
	updater  *ext.Updater
	handlers []bothandlers.Handler
	log      *slog.Logger
	cfg      *config.BotConfig
}

func New(
	cfg *config.BotConfig,
	log *slog.Logger,
	topicService bothandlers.TopicService,
	linkService bothandlers.LinkService,
) (*Bot, error) {
	bot, err := gotgbot.NewBot(cfg.Token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(_ *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Error("an error occurred while handling update", slog.Any("error", err))
			return ext.DispatcherActionNoop
		},
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{
		UnhandledErrFunc: func(_ error) {},
	})

	var handle []bothandlers.Handler
	handle = append(
		handle,
		bothandlers.NewTopicsHandler(cfg, log, topicService),
		bothandlers.NewLinksHandler(cfg, log, linkService),
	)

	for _, h := range handle {
		h.Register(dispatcher)
	}

	log.Info("bot configured successfully", slog.String("name", bot.Username))

	return &Bot{
		api:      bot,
		updater:  updater,
		handlers: handle,
		log:      log,
		cfg:      cfg,
	}, nil
}

func (b *Bot) MustRun() {
	err := b.updater.StartPolling(b.api, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout:        int64(b.cfg.UpdateTimeout.Seconds()),
			AllowedUpdates: []string{"message"},
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: b.cfg.RequestTimeout,
			},
		},
	})

	if err != nil {
		panic("failed to start polling: " + err.Error())
	}

	b.updater.Idle()
}

func (b *Bot) Stop() {
	if err := b.updater.Stop(); err != nil {
		b.log.Error("failed to stop updater", slog.Any("error", err))
	}
}
