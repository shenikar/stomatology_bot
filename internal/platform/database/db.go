package database

import (
	"context"
	"fmt"
	"stomatology_bot/configs"

	"github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v5"
)

func GetConnect(dbConfig configs.DBConfig) (*pgx.Conn, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}
	logrus.Info("Connect database")
	return conn, nil
}
