package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/dylanbr0wn/clockr/internal/db"
	"github.com/dylanbr0wn/clockr/internal/db/sqlc"
	"github.com/dylanbr0wn/clockr/internal/seed"
)

// App struct
type App struct {
	ctx     context.Context
	conn    *sql.DB
	Queries *sqlc.Queries
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved so we can call
// the runtime methods. It also opens the local database, self-migrates to the
// latest schema, and seeds core data on first run.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	conn, err := db.OpenDefault()
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := seed.Core(ctx, conn); err != nil {
		log.Fatalf("seed database: %v", err)
	}

	a.conn = conn
	a.Queries = sqlc.New(conn)
}

// shutdown is called on app exit; close the database cleanly.
func (a *App) shutdown(ctx context.Context) {
	if a.conn != nil {
		_ = a.conn.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
