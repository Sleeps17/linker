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
	"github.com/olekukonko/tablewriter"
	"log/slog"
)

const (
	postTopicCmd   = "post_topic"
	deleteTopicCmd = "delete_topic"
	listTopicsCmd  = "list_topics"
)

type TopicService interface {
	PostTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	DeleteTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	ListTopics(ctx context.Context, username string) (topics []string, err error)
}

type TopicsHandler struct {
	topicService TopicService
	log          *slog.Logger
	cfg          *config.BotConfig
}

func NewTopicsHandler(
	cfg *config.BotConfig,
	log *slog.Logger,
	topicService TopicService,
) *TopicsHandler {
	return &TopicsHandler{
		cfg:          cfg,
		log:          log,
		topicService: topicService,
	}
}

func (h *TopicsHandler) Register(dispatcher *ext.Dispatcher) {
	cmdHandlers := []handlers.Response{
		h.postTopic,
		h.deleteTopic,
		h.listTopics,
	}

	cmdTags := []string{
		postTopicCmd,
		deleteTopicCmd,
		listTopicsCmd,
	}

	for idx := range cmdHandlers {
		dispatcher.AddHandler(handlers.NewCommand(
			cmdTags[idx],
			cmdHandlers[idx],
		))
	}
}

func (h *TopicsHandler) postTopic(bot *gotgbot.Bot, extctx *ext.Context) error {
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

	id, err := h.topicService.PostTopic(ctx, username, args.Topic)
	if err != nil {
		if errors.Is(err, storage.ErrTopicAlreadyExists) {
			if err := sendMessage(bot, chatID, "Топик с таким названием уже существует"); err != nil {
				return err
			}

			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Не удалось создать топик"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	if err := sendMessage(bot, chatID, fmt.Sprintf("Топик усмпешно создан, id = %d", id)); err != nil {
		return err
	}
	return ext.EndGroups
}

func (h *TopicsHandler) deleteTopic(bot *gotgbot.Bot, extctx *ext.Context) error {
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

	id, err := h.topicService.DeleteTopic(ctx, username, args.Topic)
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

		if err := sendMessage(bot, chatID, "Не удалось удалить топик"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	if err := sendMessage(bot, chatID, fmt.Sprintf("Топик усмпешно удален, id = %d", id)); err != nil {
		return err
	}
	return ext.EndGroups
}

func (h *TopicsHandler) listTopics(bot *gotgbot.Bot, extctx *ext.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), handlersTimeout)
	defer cancel()

	chatID := extctx.Message.Chat.Id
	username := extctx.Message.From.Username

	topics, err := h.topicService.ListTopics(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			if err := sendMessage(bot, chatID, "Пользователь не найден"); err != nil {
				return err
			}
			return ext.EndGroups
		}

		if err := sendMessage(bot, chatID, "Неудалось получить список топиков"); err != nil {
			return err
		}
		return ext.EndGroups
	}

	var buffer bytes.Buffer
	table := tablewriter.NewWriter(&buffer)
	headers := []string{"ID", "Topic"}
	values := make([][]string, 0)
	for idx, topic := range topics {
		values = append(values, []string{fmt.Sprint(idx + 1), topic})
	}

	table.SetHeader(headers)
	table.AppendBulk(values)
	table.Render()

	if err := sendMessageMD(bot, chatID, buffer.String()); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return ext.EndGroups
}
