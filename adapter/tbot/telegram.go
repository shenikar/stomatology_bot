package tbot

import (
	"log"
	"stomatology_bot/config"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBot struct {
	api *tgbot.BotAPI
	cfg *config.Config
}

func NewBot(api *tgbot.BotAPI, cfg *config.Config) *TgBot {
	return &TgBot{
		api: api,
		cfg: cfg,
	}
}

func (b *TgBot) Start() {
	// Настраиваем канал для получения обновлений
	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	// Обрабатываем обновления
	for update := range updates {
		if update.Message != nil { // Если пришло сообщение
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Обрабатываем команды
			switch update.Message.Text {
			case "/start":
				msg := tgbot.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Я бот для записи в стоматологическую клинику. Используйте команду /help для получения списка доступных команд.")
				b.api.Send(msg)
			case "/help":
				msg := tgbot.NewMessage(update.Message.Chat.ID, "Список доступных команд:\n/start - Начать работу с ботом\n/help - Помощь\n/book - Записаться на приём")
				b.api.Send(msg)
			default:
				msg := tgbot.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
				b.api.Send(msg)
			}
		}
	}
}
