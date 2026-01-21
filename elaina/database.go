package elaina

import (
	"database/sql"
	"errors"
	"log/slog"

	_ "modernc.org/sqlite"
)

const database = "data/database.db"

var tables = []string{
	`
	CREATE TABLE IF NOT EXISTS Macro (
	    key TEXT NOT NULL PRIMARY KEY, 
	    response TEXT
	);
	`,
}

func GetMacro(key string) (*Macro, error) {
	row, err := queryRow("SELECT * FROM Macro WHERE key=?", key)
	if err != nil {
		return nil, err
	}

	var macro Macro
	if err = row.Scan(&macro.Key, &macro.Response); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &macro, nil
}

func SetMacro(macro Macro) error {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(`SELECT exists(SELECT 1 FROM Macro WHERE key=?) as row_exists`, macro.Key)
	exists := false
	if err = row.Scan(&exists); err != nil {
		return err
	}

	if exists {
		row = db.QueryRow(`UPDATE Macro SET response=? WHERE key=?`, macro.Response, macro.Key)
	} else {
		row = db.QueryRow(`INSERT INTO Macro (key, response) VALUES (?, ?)`, macro.Key, macro.Response)
	}

	return err
}

func DeleteMacro(key string) (int, error) {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	row, err := db.Query(`DELETE FROM Macro WHERE key=? RETURNING key`, key)
	if err != nil {
		return 0, err
	}
	defer row.Close()

	deleted := 0
	for row.Next() {
		deleted++
	}

	return deleted, nil
}

func queryRow(query string, args ...any) (*sql.Row, error) {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.QueryRow(query, args...), nil
}

func InitializeDatabase() error {
	slog.Info("Initialising database")
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, query := range tables {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	slog.Info("Database initialized")
	return nil
}
