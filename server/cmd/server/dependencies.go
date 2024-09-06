package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/skryde/booking-check/server/internal/api"
	"github.com/skryde/booking-check/server/internal/notification"
	"github.com/skryde/booking-check/server/internal/platform/queue"
	"github.com/skryde/booking-check/server/internal/platform/storage/badger"
	"github.com/skryde/booking-check/server/internal/platform/telegrambot"
)

const (
	botDescription = `This bot will try to help you getting a Montevideo's Spain Consulate booking hour by notifying you when the booking system shows hour availability.

Privacy: at the moment you /subscribe to these notifications, I will only save your Telegram user ID.

Disclaimer: this bot was not created by the Spain Consulate and is not an official communication channel of them; this is just a simple bot that will send you a message when it detects hour availability in the booking system.`
)

func buildConfiguration() (configuration, error) {
	readOSEnv := func(key string) (string, error) {
		value := os.Getenv(key)
		if value == "" {
			return "", fmt.Errorf("%s environment variable not set", key)
		}

		return value, nil
	}

	dbPath, err := readOSEnv("DB_PATH")
	if err != nil {
		return configuration{}, err
	}

	botToken, err := readOSEnv("TELEGRAM_BOT_TOKEN")
	if err != nil {
		return configuration{}, err
	}

	botOwnerID, err := readOSEnv("TELEGRAM_BOT_OWNER_ID")
	if err != nil {
		return configuration{}, err
	}

	ownerID, err := strconv.ParseInt(botOwnerID, 10, 64)
	if err != nil {
		return configuration{}, fmt.Errorf("invalid '%s' owner Telegram ID: %w", botOwnerID, err)
	}

	return configuration{
			dbPath:             dbPath,
			telegramBotToken:   botToken,
			telegramBotOwnerID: ownerID,
		},
		nil
}

func buildDependencies(ctx context.Context, _queue *queue.Queue) (dependencies, error) {
	cfg, err := buildConfiguration()
	if err != nil {
		return dependencies{}, fmt.Errorf("failed to build configuration: %w", err)
	}

	db, err := badger.NewDB(cfg.dbPath)
	if err != nil {
		return dependencies{}, fmt.Errorf("error creating database instance: %w", err)
	}

	botSubsHandler := notification.NewBotSubscriptionHandler(db, cfg.telegramBotOwnerID)
	bot, err := telegrambot.NewBot(cfg.telegramBotToken, botDescription, botSubsHandler.Start)
	if err != nil {
		return dependencies{}, fmt.Errorf("error creating telegram bot: %w", err)
	}

	err = bot.RegisterCommandHandler("/subscribe",
		"Subscribe to Spain Consulate Hour check",
		botSubsHandler.Subscribe,
	)
	if err != nil {
		return dependencies{}, fmt.Errorf("error registering command: %w", err)
	}

	err = bot.RegisterCommandHandler("/unsubscribe",
		"Unsubscribe from Spain Consulate Hour check",
		botSubsHandler.Unsubscribe,
	)
	if err != nil {
		return dependencies{}, fmt.Errorf("error registering command: %w", err)
	}

	err = bot.RegisterCommandHandler("/enabledebug", "",
		botSubsHandler.EnableDebug,
	)
	if err != nil {
		return dependencies{}, fmt.Errorf("error registering command: %w", err)
	}

	err = bot.RegisterCommandHandler("/disabledebug", "",
		botSubsHandler.DisableDebug,
	)
	if err != nil {
		return dependencies{}, fmt.Errorf("error registering command: %w", err)
	}

	err = bot.RegisterCommandHandler("/status", "",
		botSubsHandler.Status,
	)
	if err != nil {
		return dependencies{}, fmt.Errorf("error registering command: %w", err)
	}

	queueHandler := notification.NewQueueHandler(ctx, bot, db, _queue, cfg.telegramBotOwnerID)
	err = _queue.Subscribe(notification.NotifierTopicName, queueHandler.NotifyTopic)
	if err != nil {
		return dependencies{}, fmt.Errorf("error subscribing to topic '%s': %w",
			notification.NotifierTopicName, err,
		)
	}

	err = _queue.Subscribe(notification.ScrapperResultTopicName, queueHandler.ScrapperResultTopic)
	if err != nil {
		return dependencies{}, fmt.Errorf("error subscribing to topic '%s': %w",
			notification.ScrapperResultTopicName, err,
		)
	}

	deps := dependencies{
		bot: bot,
		api: api.NewHandler(db),
		tearDown: func() {
			slog.Info("tearing down services")

			err := db.Close()
			if err != nil {
				slog.Error("error closing database instance", slog.Any("error", err))
			}

			_queue.Shutdown()
		},
	}

	return deps, nil
}
