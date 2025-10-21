package main

import (
	"database/sql"
	"fmt"
	"log"
	"stomatology_bot/adapter/calendar"
	"stomatology_bot/adapter/db"
	"stomatology_bot/adapter/tbot"
	"stomatology_bot/config"
	"stomatology_bot/repository"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Db.User, cfg.Db.Password, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)

	// Применение миграций
	{
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			log.Fatalf("could not connect to the database for migration: %v", err)
		}
		driver, err := pgx.WithInstance(db, &pgx.Config{})
		if err != nil {
			log.Fatalf("could not create the pgx driver: %v", err)
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"pgx", driver)
		if err != nil {
			log.Fatalf("could not create the migrate instance: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("could not run the up migrations: %v", err)
		}
		log.Println("Migrations applied successfully")
		db.Close()
	}

	// Основное подключение к БД
	pgxConn, err := db.GetConnect(cfg.Db)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewBookingRepo(pgxConn)

	calendarSvc, err := calendar.NewCalendarService("credentials.json", cfg.Telegram.CalendarID)
	if err != nil {
		log.Fatal(err)
	}

	botApi, err := tgbot.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}
	botApi.Debug = true
	log.Printf("Authorized on account %s", botApi.Self.UserName)

	bot := tbot.NewBot(botApi, cfg, repo, calendarSvc)
	bot.Start()

}
