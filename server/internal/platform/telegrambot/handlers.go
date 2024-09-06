package telegrambot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func meHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "Your Telegram user ID is: `%d`"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      fmt.Sprintf(message, update.Message.Chat.ID),
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func defaultStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := `Welcome to this friendly bot!`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}
