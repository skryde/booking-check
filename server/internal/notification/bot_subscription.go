package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/skryde/booking-check/server/internal/repository"
)

type BotSubscriptionHandler struct {
	db repository.Repository

	telegramBotOwner int64
}

func NewBotSubscriptionHandler(db repository.Repository, telegramBotOwner int64) *BotSubscriptionHandler {
	return &BotSubscriptionHandler{
		db:               db,
		telegramBotOwner: telegramBotOwner,
	}
}

func (s *BotSubscriptionHandler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := `Welcome!

Use /subscribe command to subscribe to the hour availability notification.
Use /unsubscribe command to stop receiving notifications.
`
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

func (s *BotSubscriptionHandler) Subscribe(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageText := `User subscribed`

	err := s.db.AddSubscriber(update.Message.Chat.ID)
	if err != nil {
		slog.Error("error adding subscriber to DB",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageText = "Error subscribing to the notifications"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   messageText,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (s *BotSubscriptionHandler) Unsubscribe(ctx context.Context, b *bot.Bot, update *models.Update) {
	messageText := `User unsubscribed`

	err := s.db.RemoveSubscriber(update.Message.Chat.ID)
	if err != nil {
		slog.Error("error removing subscriber from DB",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageText = "Error unsubscribing to the notifications"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   messageText,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

//
// Debug manager handlers
//

func (s *BotSubscriptionHandler) EnableDebug(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Ignore the command if the user is not the Bot Owner.
	if update.Message.Chat.ID != s.telegramBotOwner {
		return
	}

	messageText := `Debug enabled`

	err := s.db.ManageDebug(true)
	if err != nil {
		slog.Error("error enabling debug",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageText = "Error enabling debug"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   messageText,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (s *BotSubscriptionHandler) DisableDebug(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Ignore the command if the user is not the Bot Owner.
	if update.Message.Chat.ID != s.telegramBotOwner {
		return
	}

	messageText := `Debug disabled`

	err := s.db.ManageDebug(false)
	if err != nil {
		slog.Error("error disabling debug",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageText = "Error disabling debug"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   messageText,
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}

func (s *BotSubscriptionHandler) Status(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Ignore the command if the user is not the Bot Owner.
	if update.Message.Chat.ID != s.telegramBotOwner {
		return
	}

	messageTemplate := `System status:

Debug status: %t
Subscriptions: %+v
`

	debugEnabled, err := s.db.DebugEnabled()
	if err != nil {
		slog.Error("error getting debug status",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageTemplate = "Error getting debug status"
	}

	subs, err := s.db.Subscribers()
	if err != nil {
		slog.Error("error getting subscribers",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
		messageTemplate = "Error getting subscribers"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf(messageTemplate, debugEnabled, subs),
	})
	if err != nil {
		slog.Error("error sending message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Any("error", err),
		)
	}
}
