package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Sqlite adapter
func Sqlite(dbFile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Drop the sessions table if it exists
	_, err = db.Exec(`DROP TABLE IF EXISTS sessions`)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Create the sessions table
	_, err = db.Exec(`
        CREATE TABLE sessions (
          id SERIAL PRIMARY KEY,
          user_id TEXT NOT NULL,
          expires_at TIMESTAMP NOT NULL
       );
   `)
	if err != nil {
		log.Fatalf("failed to set up sessions w table: %s", err)
	}

	// Create the sessions table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
          id SERIAL PRIMARY KEY,
          email TEXT NOT NULL,
          password TEXT NOT NULL,
          salt TEXT NOT NULL
       );
   `)

	if err != nil {
		log.Fatalf("failed to set up auth table: %s", err)
	}

	return db
}
