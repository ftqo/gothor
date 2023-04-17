package db

import (
	"database/sql"
	"fmt"

	"github.com/ftqo/gothor/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(c config.DB) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping sql database: %v", err)
	}

	return db, nil
}
