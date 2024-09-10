package app

import (
	botapp "github.com/Sleeps17/linker/internal/app/bot"
	httpapp "github.com/Sleeps17/linker/internal/app/http"
	urlShortener "github.com/Sleeps17/linker/internal/clients/url-shortener"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/storage"
	"log/slog"
)

type app interface {
	MustRun()
	Stop()
}

type Service struct {
	apps []app
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	storage storage.Storage,
	_ urlShortener.UrlShortener,
) *Service {
	var apps []app

	apps = append(
		apps,
		httpapp.New(
			&cfg.Rest,
			log,
			storage,
		),
	)

	apps = append(
		apps,
		botapp.MustNew(
			&cfg.Bot,
			log,
			storage,
		),
	)

	return &Service{
		apps: apps,
	}
}

func (s *Service) MustStart() {
	for _, a := range s.apps {
		go a.MustRun()
	}
}

func (s *Service) Stop() {
	for _, a := range s.apps {
		a.Stop()
	}
}
