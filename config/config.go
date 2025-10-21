package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram TelegramConfig
	Db       DbConfig
}
type TelegramConfig struct {
	Token      string
	CalendarID string
}
type DbConfig struct {
	User     string
	Password string
	Name     string
	Port     string
	Host     string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	telegramConfig := TelegramConfig{
		Token:      os.Getenv("BOT_TOKEN"),
		CalendarID: os.Getenv("CALENDAR_ID"),
	}
	dbConfig := DbConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
		Host:     os.Getenv("DB_HOST"),
	}
	return &Config{
		Telegram: telegramConfig,
		Db:       dbConfig,
	}, nil
}
