package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var schema = `
CREATE TABLE IF NOT EXISTS teams (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER,
	user_name TEXT UNIQUE,
	first_name TEXT NOT NULL,
	team_id INTEGER REFERENCES teams (id),
	admin BOOLEAN DEFAULT FALSE,
	PRIMARY KEY(id)
);
`

// DB is a wrapper around our Postgres database
type DB struct {
	*sqlx.DB
	logger *zap.Logger
}

// New returns a new Store instance
func New(connStr string, logger *zap.Logger) (*DB, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to postgres: %v", err)
	}

	db.MustExec(schema)

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}
