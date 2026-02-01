package common

import (
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

var guildSettingsCache = CreateCache[Snowflake, GuildSettings](5)

// ConnectDatabase attempts to open a connection with Elaina's backend database. If a connection can't be established,
// the state is assumed to be unrecoverable and panics.
func ConnectDatabase(user string, password string, address string) *sql.DB {
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

// DeployDatabase reads migration.sql files from $PWD/migrations and executes all of them which aren't present in the
// migrations.migration table. These statements are run as root, use with care.
func DeployDatabase(user string, password string, address string) {
	slog.Info("[Database] Starting database deployment...")
	conn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?multiStatements=true", user, password, address))
	AssertIsNil(err)
	defer conn.Close()

	_, err = conn.Exec(`CREATE SCHEMA IF NOT EXISTS migrations;CREATE SCHEMA IF NOT EXISTS elaina;USE elaina;`)
	AssertIsNil(err)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS migrations.migration(version VARCHAR(10) PRIMARY KEY);`)
	AssertIsNil(err)

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
		AssertIsNil(err)

		slog.Info("[Database] Running migration script: " + name)
		tx, err := conn.Begin() // Run the migration as a transaction
		AssertIsNil(err)

		_, err = tx.Exec(string(script))
		AssertIsNil(err)
		_, err = tx.Exec("INSERT INTO migrations.migration (version) VALUES (?)", ver)
		AssertIsNil(err)

		AssertIsNil(tx.Commit())
	}
	slog.Info("[Database] Finished database deployment")
}

// GetGuildSettings fetches the settings for the given guild and returns them. If no settings are found in the cache or
// database, the default settings are returned instead.
func GetGuildSettings(guild Snowflake) (GuildSettings, error) {
	settings, err := fetchGuildSettings(guild)
	if err != nil {
		return GuildSettings{}, err
	} else if settings != nil {
		return *settings, nil
	}

	return DefaultGuildSettings(), nil
}

func CreateOrUpdateGuildSettings(guild Snowflake, settings GuildSettings) error {
	_, err := dbConn.Exec(`INSERT INTO guild_settings VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE honeypot_channel=values(honeypot_channel), hello_enabled=values(hello_enabled)`,
		guild, settings.HoneypotChannel, settings.HelloEnabled)
	guildSettingsCache.Add(guild, settings)
	return err
}

func fetchGuildSettings(guild Snowflake) (*GuildSettings, error) {
	if val := guildSettingsCache.Get(guild); val != nil {
		return val, nil
	}
	row := dbConn.QueryRow(`SELECT honeypot_channel, hello_enabled FROM guild_settings WHERE guild_id=?`, guild)

	var settings GuildSettings
	if err := row.Scan(&settings.HoneypotChannel, &settings.HelloEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

func GetMacro(guild Snowflake, keyword string) (*Macro, error) {
	row := dbConn.QueryRow(`SELECT * FROM macro WHERE guild_id=? AND keyword=?`, guild, keyword)

	var macro Macro
	if err := row.Scan(&macro.Guild, &macro.Key, &macro.Response); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &macro, nil
}

func CreateOrUpdateMacro(macro Macro) error {
	_, err := dbConn.Exec(`INSERT INTO macro VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE response=values(response)`,
		macro.Guild, macro.Key, macro.Response)
	return err
}

func DeleteMacro(guild Snowflake, key string) (bool, error) {
	res, err := dbConn.Exec(`DELETE FROM macro WHERE guild_id=? AND keyword=?`, guild, key)
	if err != nil {
		return false, err
	}
	i, err := res.RowsAffected()
	return i > 0, err
}
