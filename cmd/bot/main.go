package main

import (
	"database/sql"
	"fmt"
	"stomatology_bot/configs"
	"stomatology_bot/internal/booking"
	"stomatology_bot/internal/logger"
	"stomatology_bot/internal/platform/calendar"
	"stomatology_bot/internal/platform/database"
	"stomatology_bot/internal/platform/telegram"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
	}
	logger.SetupLogger(cfg.LogLevel)
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)

	// Применение миграций
	{
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			logrus.WithError(err).Fatal("could not connect to the database for migration")
		}
		driver, err := pgx.WithInstance(db, &pgx.Config{})
		if err != nil {
			logrus.WithError(err).Fatal("could not create the pgx driver")
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"pgx", driver)
		if err != nil {
			logrus.WithError(err).Fatal("could not create the migrate instance")
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logrus.WithError(err).Fatal("could not run the up migrations")
		}
		logrus.Info("Migrations applied successfully")
		db.Close()
	}

	// Основное подключение к БД
	pgxConn, err := database.GetConnect(cfg.DB)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	repo := booking.NewRepo(pgxConn)

	calendarSvc, err := calendar.NewService("credentials.json", cfg.Telegram.CalendarID, cfg.Telegram.WorkStartHour, cfg.Telegram.WorkEndHour)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create calendar service")
	}

	botAPI, err := tgbot.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create bot API")
	}
	botAPI.Debug = true
	logrus.Infof("Authorized on account %s", botAPI.Self.UserName)

	bot := telegram.NewBot(botAPI, cfg, repo, calendarSvc)
	bot.Start()
}
