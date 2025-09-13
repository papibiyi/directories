package models

import (
    "database/sql"
    "log"
    "sync"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
    Db   *sql.DB
    once sync.Once
}

// InitializeDB initializes the database connection and creates the table if it doesn't exist.
func (a *App) InitializeDB() {
    a.once.Do(func() {
        db, err := sql.Open("sqlite3", "directory_app.db")
        if err != nil {
            log.Fatal(err)
        }
        a.Db = db

        createTable := `
		
	CREATE TABLE IF NOT EXISTS directory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		phone_number TEXT,
		created_at TEXT,
		updated_at TEXT
	);

	CREATE TABLE IF NOT EXISTS address (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		directory_id INTEGER NOT NULL,
		address_line_1 TEXT,
		address_line_2 TEXT,
		city TEXT,
		state TEXT,
		country TEXT,
		FOREIGN KEY (directory_id) REFERENCES directory (id) ON DELETE CASCADE
	);
		
		`
        _, err = a.Db.Exec(createTable)
        if err != nil {
            log.Fatalf("Could not create table: %s\n", err.Error())
        }
    })
}
