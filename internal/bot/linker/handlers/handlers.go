package bothandlers

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/Sleeps17/linker/internal/models"
	"regexp"
	"strings"
	"time"
)

const (
	handlersTimeout = 5 * time.Second

	commandPattern = `^\/(?P<command>\w+)(?:\s+(topic:(?P<topic>[^ ]+)|link:(?P<link>[^ ]+)|alias:(?P<alias>[^ ]+)))*$`
)

type Handler interface {
	Register(dispatcher *ext.Dispatcher)
}

func sendMessage(api *gotgbot.Bot, chatID int64, text string) error {
	_, err := api.SendMessage(chatID, text, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func sendMessageMD(api *gotgbot.Bot, chatID int64, text string) error {
	escapedText := escapeMarkdownV2(text)
	escapedText = "```\n" + escapedText + "\n```"

	_, err := api.SendMessage(chatID, escapedText, &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func parseCommandArgs(msg string) (*models.CmdArgs, error) {
	re, err := regexp.Compile(commandPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regexp: %w", err)
	}

	match := re.FindStringSubmatch(msg)
	if match == nil {
		return nil, fmt.Errorf("failed to parse command")
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return &models.CmdArgs{
		Topic: result["topic"],
		Link:  result["link"],
		Alias: result["alias"],
	}, nil
}

func escapeMarkdownV2(text string) string {
	// Экранирование всех специальных символов для MarkdownV2
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
