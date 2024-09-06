package notification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/skryde/booking-check/server/internal/platform/telegrambot"
	"github.com/skryde/booking-check/server/internal/repository"
)

const (
	ScrapperResultTopicName = "scrapper.result"
	NotifierTopicName       = "notify"
)

type Publisher interface {
	Publish(string, []byte) error
}

type QueueHandler struct {
	ctx context.Context

	bot *telegrambot.TelegramBot
	db  repository.Repository

	publisher Publisher

	telegramBotOwner int64
}

func NewQueueHandler(
	ctx context.Context,
	bot *telegrambot.TelegramBot,
	db repository.Repository,
	publisher Publisher,
	telegramBotOwner int64,
) *QueueHandler {
	return &QueueHandler{
		ctx:              ctx,
		bot:              bot,
		db:               db,
		publisher:        publisher,
		telegramBotOwner: telegramBotOwner,
	}
}

func (q *QueueHandler) ScrapperResultTopic(m *nats.Msg) {
	var payload struct {
		Debug   bool   `json:"debug"`
		Message string `json:"message"`
		Image   string `json:"image"`
	}

	err := json.Unmarshal(m.Data, &payload)
	if err != nil {
		slog.Error("error unmarshalling message", slog.String("payload", string(m.Data)))
		return
	}

	if payload.Debug {
		debugEnabled, err := q.db.DebugEnabled()
		if err != nil {
			slog.Error("error getting debug status",
				slog.Any("error", err),
			)
			return // On error assume debug=false.
		}

		if !debugEnabled {
			return
		}

		err = q.publish(
			q.telegramBotOwner,
			payload.Message,
			payload.Image,
		)
		if err != nil {
			slog.Error("error publishing message",
				slog.String("destiny_topic", NotifierTopicName),
				slog.Int64("recipient", q.telegramBotOwner),
				slog.String("message", payload.Message),
				slog.Any("error", err),
			)
			return
		}

		return
	}

	subs, err := q.db.Subscribers()
	if err != nil {
		slog.Error("error getting subscribers")
		return
	}

	for _, subscriber := range subs {
		err := q.publish(subscriber, payload.Message, payload.Image)
		if err != nil {
			slog.Error("error publishing message",
				slog.String("destiny_topic", NotifierTopicName),
				slog.Int64("recipient", subscriber),
				slog.String("message", payload.Message),
				slog.Any("error", err),
			)

			continue
		}
	}
}

func (q *QueueHandler) publish(recipient int64, message, image string) error {
	var notification struct {
		Recipient int64  `json:"recipient"`
		Message   string `json:"message"`
		Image     string `json:"image"`
	}
	notification.Recipient = recipient
	notification.Message = message
	notification.Image = image

	b, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	err = q.publisher.Publish(NotifierTopicName, b)
	if err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}

	return nil
}

func (q *QueueHandler) NotifyTopic(m *nats.Msg) {
	var payload struct {
		Recipient int64  `json:"recipient"`
		Message   string `json:"message"`
		Image     string `json:"image"`
	}

	err := json.Unmarshal(m.Data, &payload)
	if err != nil {
		slog.Error("error unmarshalling message",
			slog.Any("error", err),
			slog.String("topic_name", m.Subject),
			slog.Int64("recipient", payload.Recipient),
			slog.String("message", payload.Message),
		)
		return
	}

	err = q.bot.SendMessage(q.ctx, payload.Recipient, payload.Message)
	if err != nil {
		slog.Error("error sending text message message",
			slog.Any("error", err),
			slog.String("topic_name", m.Subject),
			slog.Int64("recipient", payload.Recipient),
			slog.String("message", payload.Message),
		)
		return
	}

	img, err := base64.StdEncoding.DecodeString(payload.Image)
	if err != nil {
		slog.Error("error decoding base64 image",
			slog.Any("error", err),
			slog.String("topic_name", m.Subject),
			slog.Int64("recipient", payload.Recipient),
			slog.String("message", payload.Message),
		)
		return
	}

	err = q.bot.SendPhoto(q.ctx, payload.Recipient, img)
	if err != nil {
		slog.Error("error unmarshalling message",
			slog.Any("error", err),
			slog.String("topic_name", m.Subject),
			slog.Int64("recipient", payload.Recipient),
			slog.String("message", payload.Message),
		)
		return
	}
}
