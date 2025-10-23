package configs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram TelegramConfig
	DB       DBConfig
	LogLevel string
}
type TelegramConfig struct {
	Token         string
	CalendarID    string
	AdminID       string
	WorkStartHour int
	WorkEndHour   int
}
type DBConfig struct {
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
		Token:         os.Getenv("BOT_TOKEN"),
		CalendarID:    os.Getenv("CALENDAR_ID"),
		AdminID:       os.Getenv("ADMIN_ID"),
		WorkStartHour: parseInt(os.Getenv("WORK_START_HOUR"), 9), // Значение по умолчанию 9
		WorkEndHour:   parseInt(os.Getenv("WORK_END_HOUR"), 18),  // Значение по умолчанию 18
	}
	dbConfig := DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
		Host:     os.Getenv("DB_HOST"),
	}
	return &Config{
		Telegram: telegramConfig,
		DB:       dbConfig,
		LogLevel: os.Getenv("LOG_LEVEL"),
	}, nil
}

// parseInt пытается преобразовать строку в int, возвращая defaultValue в случае ошибки
func parseInt(s string, defaultValue int) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
