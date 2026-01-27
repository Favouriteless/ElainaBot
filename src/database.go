package main

import (
	"ElainaBot/discord"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const dbConnectAttempts = 5

var dbConn *sql.DB

func connectDatabase(user string, password string, address string) *sql.DB {
	slog.Info("[Database] Connecting to database")

	var err error

	wait := 3
	for i := 0; i < dbConnectAttempts; i++ {
		dbConn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/elaina", user, password, address))
		if err == nil {
			break
		}
		slog.Warn(fmt.Sprintf("[Database] Failed to connect to database. Waiting %d seconds", wait), slog.Int("attempt", i))
		time.Sleep(time.Second * time.Duration(wait))
		wait *= wait // Exponential backoff
	}

	if err != nil || dbConn == nil {
		slog.Error("[Database] Failed to connect to database, terminating process...")
		panic(err) // Failure to connect is an unrecoverable state
	}

	slog.Info("[Database] Database connection established")
	return dbConn
}

// deployDatabase reads migration.sql files from $PWD/migrations and executes all of them which aren't present in the
// migrations.migration table. These statements are run as root, use with care.
func deployDatabase(user string, password string, address string) {
	slog.Info("[Database] Starting database deployment...")
	conn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?multiStatements=true", user, password, address))
	assertIsNil(err)
	defer conn.Close()

	_, err = conn.Exec(`CREATE SCHEMA IF NOT EXISTS migrations;CREATE SCHEMA IF NOT EXISTS elaina;USE elaina;`)
	assertIsNil(err)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS migrations.migration(version VARCHAR(10) PRIMARY KEY);`)
	assertIsNil(err)

	files, err := os.ReadDir("migrations")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		ver := name[:strings.IndexByte(name, '_')]

		res := conn.QueryRow(`SELECT * FROM migrations.migration WHERE version=?`, ver)

		var found string
		if err = res.Scan(&found); err == nil {
			continue // We don't want to apply migrations we have already applied
		} else if !errors.Is(err, sql.ErrNoRows) {
			panic(err) // ErrNoRows is expected, but other errors are fatal here
		}

		script, err := os.ReadFile("migrations/" + name)
		assertIsNil(err)

		tx, err := conn.Begin() // Run the migration as a transaction
		assertIsNil(err)

		_, err = tx.Exec(string(script))
		assertIsNil(err)
		_, err = tx.Exec("INSERT INTO migrations.migration (version) VALUES (?)", ver)
		assertIsNil(err)

		assertIsNil(tx.Commit())
	}
	slog.Info("[Database] Finished database deployment")
}

func GetMacro(guild discord.Snowflake, keyword string) (*Macro, error) {
	row := dbConn.QueryRow("SELECT * FROM macro WHERE guild=? AND keyword=?", guild, keyword)

	var macro Macro
	if err := row.Scan(&macro.Guild, &macro.Key, &macro.Response); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &macro, nil
}

func CreateBan(guild discord.Snowflake, user discord.Snowflake, expires int64, reason string) (err error) {
	exists := false
	if err = dbConn.QueryRow(`SELECT exists(SELECT 1 FROM ban WHERE guild=? AND user_id=?)`, guild, user).Scan(&exists); err != nil {
		return err
	}

	if exists {
		_, err = dbConn.Exec(`UPDATE ban SET expires=?, reason=? WHERE guild=? AND user_id=?`, expires, reason, user)
	} else {
		_, err = dbConn.Exec(`INSERT INTO ban (guild, user_id, expires, reason) VALUES (?, ?, ?, ?)`, guild, user, expires, reason)
	}
	return err
}

func DeleteBan(guild discord.Snowflake, user discord.Snowflake) error {
	_, err := deleteRow(`DELETE FROM ban WHERE guild=? AND user_id=?`, guild, user)
	return err
}

func GetExpiredBans(time int64) ([]Ban, error) {
	rows, err := dbConn.Query(`SELECT * FROM ban WHERE expires<=?`, time)
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

func SetMacro(macro Macro) (err error) {
	exists := false
	if err = dbConn.QueryRow(`SELECT exists(SELECT 1 FROM macro WHERE guild=? AND keyword=?)`, macro.Guild, macro.Key).Scan(&exists); err != nil {
		return err
	}

	if exists {
		_, err = dbConn.Exec(`UPDATE macro SET response=? WHERE keyword=?`, macro.Response, macro.Key)
	} else {
		_, err = dbConn.Exec(`INSERT INTO macro (guild, keyword, response) VALUES (?, ?, ?)`, macro.Guild, macro.Key, macro.Response)
	}
	return err
}

func DeleteMacro(guild discord.Snowflake, key string) (bool, error) {
	i, err := deleteRow(`DELETE FROM macro WHERE guild=? AND keyword=?`, guild, key)
	return i > 0, err
}

func deleteRow(query string, args ...interface{}) (int64, error) {
	res, err := dbConn.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return i, nil
}
