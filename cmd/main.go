package main

import (
	"log"
	"stomatology_bot/adapter/tbot"
	"stomatology_bot/config"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	botApi, err := tgbot.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}
	botApi.Debug = true
	log.Printf("Authorized on account %s", botApi.Self.UserName)

	bot := tbot.NewBot(botApi, cfg)
	bot.Start()

}
