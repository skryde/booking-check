package telegrambot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var commandValidationPattern = regexp.MustCompile("^/[a-z]+$")

type TelegramBot struct {
	bot *bot.Bot

	myDescription string
	myCommands    bot.SetMyCommandsParams

	startHandler bot.HandlerFunc
}

// TODO: Sanitize errors to mask the Bot API Token.
func NewBot(token, myDescription string, startHandler bot.HandlerFunc) (*TelegramBot, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {}),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating new telegram bot: %w", err)
	}

	tb := &TelegramBot{
		bot:           b,
		myDescription: myDescription,
		myCommands:    bot.SetMyCommandsParams{Scope: &models.BotCommandScopeDefault{}},
		startHandler:  startHandler,
	}

	err = tb.configure()
	if err != nil {
		return nil, fmt.Errorf("error configuring telegram bot: %w", err)
	}

	return tb, nil
}

func (t *TelegramBot) configure() error {
	if err := t.RegisterCommandHandler("/me", "Returns your user's Telegram ID", meHandler); err != nil {
		return fmt.Errorf("error registering '%s' comand", "/me")
	}

	startHandler := t.startHandler
	if startHandler == nil {
		startHandler = defaultStartHandler
	}

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	return nil
}

func (t *TelegramBot) RegisterCommandHandler(pattern, description string, handler bot.HandlerFunc) error {
	if !commandValidationPattern.MatchString(pattern) {
		return errors.New("invalid command pattern: it must start with a slash and contains only minuscules letters")
	}

	// If no description is provided, assume that the dev does not want to show the command to the users.
	if len(description) > 0 {
		t.myCommands.Commands = append(t.myCommands.Commands, models.BotCommand{
			Command:     pattern[1:],
			Description: description,
		})
	}

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, pattern, bot.MatchTypeExact, handler)

	return nil
}

func (t *TelegramBot) Start(ctx context.Context) error {
	_, err := t.bot.SetMyDescription(ctx, &bot.SetMyDescriptionParams{
		Description: t.myDescription,
	})
	if err != nil {
		return fmt.Errorf("error setting telegram bot description: %w", err)
	}

	success, err := t.bot.SetMyCommands(ctx, &t.myCommands)
	if err != nil {
		return fmt.Errorf("error setting telegram bot commands: %w", err)
	}

	if !success {
		return fmt.Errorf("telegram bot commands not set")
	}

	t.bot.Start(ctx)
	return nil
}

func (t *TelegramBot) SendMessage(ctx context.Context, recipient int64, message string) error {
	_, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: recipient,
		Text:   message,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", recipient),
			slog.Any("error", err),
		)
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (t *TelegramBot) SendPhoto(ctx context.Context, recipient int64, message []byte) error {
	_, err := t.bot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: recipient,
		Photo: &models.InputFileUpload{
			Filename: "scrapper-screenshot",
			Data:     bytes.NewReader(message),
		},
	})

	if err != nil {
		slog.Error("error sending photo",
			slog.Int64("chat_id", recipient),
			slog.Any("error", err),
		)
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}
