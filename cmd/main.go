package main

import (
	"log"
	"stomatology_bot/adapter/db"
	"stomatology_bot/adapter/tbot"
	"stomatology_bot/config"
	"stomatology_bot/repository"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := db.GetConnect(cfg.Db)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewBookingRepo(db)

	botApi, err := tgbot.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}
	botApi.Debug = true
	log.Printf("Authorized on account %s", botApi.Self.UserName)

	bot := tbot.NewBot(botApi, cfg, repo)
	bot.Start()

}
