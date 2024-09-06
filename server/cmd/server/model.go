package main

import (
	"github.com/skryde/booking-check/server/internal/api"
	"github.com/skryde/booking-check/server/internal/platform/telegrambot"
)

type configuration struct {
	dbPath             string
	telegramBotToken   string
	telegramBotOwnerID int64
}

type dependencies struct {
	bot *telegrambot.TelegramBot
	api *api.Handler

	tearDown func()
}
