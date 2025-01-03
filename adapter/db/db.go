package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"stomatology_bot/config"

	"github.com/jackc/pgx/v5"
)

func GetConnect(dbConfig config.DbConfig) (*pgx.Conn, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	log.Println("Connect database")
	return conn, nil

}
