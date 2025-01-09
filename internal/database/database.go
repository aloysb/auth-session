package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const DEFAULT_SQLITE_PATH = "sessions.db"

func Init() (*sql.DB, error) {
	db_url := os.Getenv("DATABASE_URL")
	db_type := os.Getenv("DATABASE_TYPE")

	if len(db_type) == 0 {
		db_type = "sqlite"
		slog.Info("No database type specified, defaulting to sqlite")
	}

	if len(db_url) == 0 {
		slog.Info("No database url specified, defaulting to sqlite file 'sessions.db'")
		db_url = DEFAULT_SQLITE_PATH
	}

	if strings.ToLower(db_type) == "sqlite" {
		database := SQLite{path: db_url}
		return database.Init(), nil
	}

	err := "Unknown database type: " + db_type
	return nil, fmt.Errorf(err)
}
