package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetupLogger(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel // Уровень по умолчанию
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(level)
}
