package main

import (
	"ElainaBot/discord"
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
);`,
	`
CREATE TABLE IF NOT EXISTS Ban (
    guild INTEGER NOT NULL,
    user INTEGER NOT NULL,
    expires INTEGER,
    reason TEXT,
    PRIMARY KEY(guild, user)
);`,
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

func GetBan(guild discord.Snowflake, user discord.Snowflake) (*Ban, error) {
	row, err := queryRow("SELECT * FROM Ban WHERE guild=? AND user=?", guild, user)
	if err != nil {
		return nil, err
	}

	var ban Ban
	if err = row.Scan(&ban.User, &ban.Expires, &ban.Reason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ban, nil
}

func CreateBan(guild discord.Snowflake, user discord.Snowflake, expires int64, reason string) error {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return err
	}
	defer db.Close()

	exists := false
	if err = db.QueryRow(`SELECT exists(SELECT 1 FROM Ban WHERE guild=? AND user=?)`, guild, user).Scan(&exists); err != nil {
		return err
	}

	if exists {
		db.QueryRow(`UPDATE Ban SET expires=?, reason=? WHERE guild=? AND user=?`, expires, reason, user)
	} else {
		db.QueryRow(`INSERT INTO Ban (guild, user, expires, reason) VALUES (?, ?, ?, ?)`, guild, user, expires, reason)
	}
	return err
}

func DeleteBan(guild discord.Snowflake, user discord.Snowflake) error {
	_, err := deleteRow(`DELETE FROM Ban WHERE guild=? AND user=?`, guild, user)
	return err
}

func GetExpiredBans(time int64) ([]Ban, error) {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM Ban WHERE expires<=?`, time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bans []Ban
	for rows.Next() {
		var ban Ban
		if err = rows.Scan(&ban.Guild, &ban.User, &ban.Expires, &ban.Reason); err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	return bans, nil
}

func SetMacro(macro Macro) error {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return err
	}
	defer db.Close()

	exists := false
	if err = db.QueryRow(`SELECT exists(SELECT 1 FROM Macro WHERE key=?)`, macro.Key).Scan(&exists); err != nil {
		return err
	}

	if exists {
		db.QueryRow(`UPDATE Macro SET response=? WHERE key=?`, macro.Response, macro.Key)
	} else {
		db.QueryRow(`INSERT INTO Macro (key, response) VALUES (?, ?)`, macro.Key, macro.Response)
	}

	return err
}

func DeleteMacro(key string) (bool, error) {
	i, err := deleteRow(`DELETE FROM Macro WHERE key=?`, key)
	return i > 0, err
}

func deleteRow(query string, args ...interface{}) (int64, error) {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return i, nil
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
