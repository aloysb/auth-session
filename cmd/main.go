package main

import (
	"github.com/aloysb/auth-session/internal/auth"
	"github.com/aloysb/auth-session/internal/database"
	"github.com/aloysb/auth-session/internal/server"
	"github.com/aloysb/auth-session/internal/session"
)

func main() {
	db, err := database.Init()
	if err != nil {
		panic(err)
	}
	// Ensure the database connection is closed when the program exits
	defer db.Close()
	sessionService := session.New(db)
	basicAuthService := auth.New(db)
	srv := server.New(sessionService, basicAuthService)
	srv.Start(8080)
}
