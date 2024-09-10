package bothandlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/Sleeps17/linker/pkg/random"
	"github.com/olekukonko/tablewriter"
	"log/slog"
)

const (
	postLinkCmd   = "post_link"
	pickLinkCmd   = "pick_link"
	deleteLinkCmd = "delete_link"
	listLinksCmd  = "list_links"
)

type LinkService interface {
	PostLink(ctx context.Context, username, topic, link, alias string) (err error)
	PickLink(ctx context.Context, username, topic, alias string) (link string, err error)
	DeleteLink(ctx context.Context, username, topic, alias string) (err error)
	ListLinks(ctx context.Context, username, topic string) (links []string, aliases []string, err error)
}

type LinksHandler struct {
	linkService LinkService
	log         *slog.Logger
	cfg         *config.BotConfig
}

func NewLinksHandler(
	cfg *config.BotConfig,
	log *slog.Logger,
	linkService LinkService,
) *LinksHandler {
	return &LinksHandler{
		cfg:         cfg,
		log:         log,
		linkService: linkService,
	}
}

func (h *LinksHandler) Register(dispatcher *ext.Dispatcher) {
	cmdHandlers := []handlers.Response{
		h.postLink,
		h.pickLink,
		h.deleteLink,
		h.listLinks,
	}

	cmdTags := []string{
		postLinkCmd,
		pickLinkCmd,
		deleteLinkCmd,
		listLinksCmd,
	}

	for idx := range cmdHandlers {
		dispatcher.AddHandler(handlers.NewCommand(
			cmdTags[idx],
			cmdHandlers[idx],
		))
	}
}

func (h *LinksHandler) postLink(bot *gotgbot.Bot, extctx *ext.Context) error {
	h.log.Info("postLink", slog.String("chatID", fmt.Sprintf("%d", extctx.Message.Chat.Id)))
	ctx, cancel := context.WithTimeout(context.Background(), handlersTimeout)
	defer cancel()

	chatID := extctx.Message.Chat.Id

	args, err := parseCommandArgs(extctx.Message.Text)
	if err != nil {
		return fmt.Errorf("failed to parse command args: %w", err)
	}

	if args.Topic == "" {
		if err := sendMessage(bot, chatID, "Аргумент topic обязателен"); err != nil {
			return err
		}

		return ext.EndGroups
	}

	if args.Link == "" {
		if err := sendMessage(bot, chatID, "Аргумент link обязателен"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	if args.Alias == "" {
		args.Alias = random.Alias()
	}

	username := extctx.Message.From.Username
	if err := h.linkService.PostLink(
		ctx, username,
		args.Topic, args.Link,
		args.Alias,
	); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			if err := sendMessage(bot, chatID, "Пользователь не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			if err := sendMessage(bot, chatID, "Топик не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrAliasAlreadyExists) {
			if err := sendMessage(bot, chatID, "Ссылка с таким алиасом уже существует"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Не удалось добавить ссылку"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	if err := sendMessage(bot, chatID, "Ссылка усмпешно добавлена"); err != nil {
		return err
	}
	return ext.EndGroups
}

func (h *LinksHandler) pickLink(bot *gotgbot.Bot, extctx *ext.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), handlersTimeout)
	defer cancel()

	chatID := extctx.Message.Chat.Id

	args, err := parseCommandArgs(extctx.Message.Text)
	if err != nil {
		return fmt.Errorf("failed to parse command args: %w", err)
	}

	if args.Topic == "" {
		if err := sendMessage(bot, chatID, "Аргумент topic обязателен"); err != nil {
			return err
		}

		return ext.EndGroups
	}

	if args.Alias == "" {
		if err := sendMessage(bot, chatID, "Аргумент alias обязателен"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	username := extctx.Message.From.Username
	link, err := h.linkService.PickLink(ctx, username, args.Topic, args.Alias)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			if err := sendMessage(bot, chatID, "Пользователь не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			if err := sendMessage(bot, chatID, "Топик не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			if err := sendMessage(bot, chatID, "Ссылка не найдена"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Не удалось получить ссылку"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	if err := sendMessage(bot, chatID, link); err != nil {
		return err
	}
	return ext.EndGroups
}

func (h *LinksHandler) deleteLink(bot *gotgbot.Bot, extctx *ext.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), handlersTimeout)
	defer cancel()

	chatID := extctx.Message.Chat.Id

	args, err := parseCommandArgs(extctx.Message.Text)
	if err != nil {
		return fmt.Errorf("failed to parse command args: %w", err)
	}

	if args.Topic == "" {
		if err := sendMessage(bot, chatID, "Аргумент topic обязателен"); err != nil {
			return err
		}

		return ext.EndGroups
	}

	if args.Alias == "" {
		if err := sendMessage(bot, chatID, "Аргумент alias обязателен"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	username := extctx.Message.From.Username
	if err := h.linkService.DeleteLink(ctx, username, args.Topic, args.Alias); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			if err := sendMessage(bot, chatID, "Пользователь не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			if err := sendMessage(bot, chatID, "Топик не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			if err := sendMessage(bot, chatID, "Ссылка не найдена"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Не удалось удалить ссылку"); err != nil {
			return err
		}
	}

	if err := sendMessage(bot, chatID, "Ссылка усмпешно удалена"); err != nil {
		return err
	}
	return ext.EndGroups
}

func (h *LinksHandler) listLinks(bot *gotgbot.Bot, extctx *ext.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), handlersTimeout)
	defer cancel()

	chatID := extctx.Message.Chat.Id

	args, err := parseCommandArgs(extctx.Message.Text)
	if err != nil {
		return fmt.Errorf("failed to parse command args: %w", err)
	}

	if args.Topic == "" {
		if err := sendMessage(bot, chatID, "Аргумент topic обязателен"); err != nil {
			return err
		}

		return ext.EndGroups
	}

	username := extctx.Message.From.Username
	links, aliases, err := h.linkService.ListLinks(ctx, username, args.Topic)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			if err := sendMessage(bot, chatID, "Пользователь не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			if err := sendMessage(bot, chatID, "Топик не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Не удалось получить ссылки"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	var buffer bytes.Buffer
	table := tablewriter.NewWriter(&buffer)
	headers := []string{"id", "alias", "link"}
	values := make([][]string, 0)
	for idx := range links {
		values = append(values, []string{fmt.Sprint(idx + 1), aliases[idx], links[idx]})
	}

	table.SetHeader(headers)
	table.AppendBulk(values)
	table.Render()

	if err := sendMessageMD(bot, chatID, buffer.String()); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return ext.EndGroups
}
