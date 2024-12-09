package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram TelegramConfig
}
type TelegramConfig struct {
	Token string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	telegramConfig := TelegramConfig{
		Token: os.Getenv("BOT_TOKEN"),
	}
	return &Config{
		Telegram: telegramConfig,
	}, nil
}
