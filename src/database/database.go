package database

import (
	"ElainaBot/discord"
	"ElainaBot/elaina"
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

var guildSettingsCache = discord.CreateCache[discord.Snowflake, elaina.GuildSettings](5)

// Connect attempts to open a connection with Elaina's backend database. If a connection can't be established,
// the state is assumed to be unrecoverable and panics.
func Connect(user string, password string, address string) *sql.DB {
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

// Deploy reads migration.sql files from $PWD/migrations and executes all of them which aren't present in the
// migrations.migration table. These statements are run as root, use with care.
func Deploy(user string, password string, address string) {
	slog.Info("[Database] Starting database deployment...")
	conn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?multiStatements=true", user, password, address))
	elaina.AssertIsNil(err)
	defer conn.Close()

	_, err = conn.Exec(`CREATE SCHEMA IF NOT EXISTS migrations;CREATE SCHEMA IF NOT EXISTS elaina;USE elaina;`)
	elaina.AssertIsNil(err)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS migrations.migration(version VARCHAR(10) PRIMARY KEY);`)
	elaina.AssertIsNil(err)

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
		elaina.AssertIsNil(err)

		slog.Info("[Database] Running migration script: " + name)
		tx, err := conn.Begin() // Run the migration as a transaction
		elaina.AssertIsNil(err)

		_, err = tx.Exec(string(script))
		elaina.AssertIsNil(err)
		_, err = tx.Exec("INSERT INTO migrations.migration (version) VALUES (?)", ver)
		elaina.AssertIsNil(err)

		elaina.AssertIsNil(tx.Commit())
	}
	slog.Info("[Database] Finished database deployment")
}

// GetGuildSettings fetches the settings for the given guild and returns them. If no settings are found in the cache or
// database, the default settings are returned instead.
func GetGuildSettings(guild discord.Snowflake) (elaina.GuildSettings, error) {
	settings, err := fetchGuildSettings(guild)
	if err != nil {
		return elaina.GuildSettings{}, err
	} else if settings != nil {
		return *settings, nil
	}

	return elaina.DefaultGuildSettings(), nil
}

func CreateOrUpdateGuildSettings(guild discord.Snowflake, settings elaina.GuildSettings) error {
	_, err := dbConn.Exec(`INSERT INTO guild_settings VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE honeypot_channel=values(honeypot_channel), hello_enabled=values(hello_enabled)`,
		guild, settings.HoneypotChannel, settings.HelloEnabled)
	guildSettingsCache.Add(guild, settings)
	return err
}

func fetchGuildSettings(guild discord.Snowflake) (*elaina.GuildSettings, error) {
	if val := guildSettingsCache.Get(guild); val != nil {
		return val, nil
	}
	row := dbConn.QueryRow(`SELECT honeypot_channel, hello_enabled FROM guild_settings WHERE guild_id=?`, guild)

	var settings elaina.GuildSettings
	if err := row.Scan(&settings.HoneypotChannel, &settings.HelloEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

func CreateOrUpdateBan(guild discord.Snowflake, user discord.Snowflake, expires int64, reason string) error {
	_, err := dbConn.Exec(`INSERT INTO ban VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE expires=values(expires), reason=values(reason)`,
		guild, user, expires, reason)
	return err
}

func DeleteBan(guild discord.Snowflake, user discord.Snowflake) error {
	_, err := dbConn.Exec(`DELETE FROM ban WHERE guild_id=? AND user_id=?`, guild, user)
	return err
}

func GetExpiredBans(time int64) ([]elaina.Ban, error) {
	rows, err := dbConn.Query(`SELECT * FROM ban WHERE expires<=?`, time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bans []elaina.Ban
	for rows.Next() {
		var ban elaina.Ban
		if err = rows.Scan(&ban.Guild, &ban.User, &ban.Expires, &ban.Reason); err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	return bans, nil
}

func GetMacro(guild discord.Snowflake, keyword string) (*elaina.Macro, error) {
	row := dbConn.QueryRow(`SELECT * FROM macro WHERE guild_id=? AND keyword=?`, guild, keyword)

	var macro elaina.Macro
	if err := row.Scan(&macro.Guild, &macro.Key, &macro.Response); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &macro, nil
}

func CreateOrUpdateMacro(macro elaina.Macro) error {
	_, err := dbConn.Exec(`INSERT INTO macro VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE response=values(response)`,
		macro.Guild, macro.Key, macro.Response)
	return err
}

func DeleteMacro(guild discord.Snowflake, key string) (bool, error) {
	res, err := dbConn.Exec(`DELETE FROM macro WHERE guild_id=? AND keyword=?`, guild, key)
	if err != nil {
		return false, err
	}
	i, err := res.RowsAffected()
	return i > 0, err
}
