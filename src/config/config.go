package config

import (
	"ElainaBot/discord"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"sync"
)

const configPath = "data/config.json"

const (
	HelloEmoji        = "hello_emoji"
	DefaultHelloEmoji = "default_hello_emoji"
)

func defaultValues() map[string]any {
	return map[string]any{
		HelloEmoji:        "elainastare:1462289034188689468",
		DefaultHelloEmoji: "elainastare:1462289034188689468",
	}
}

var config = &Config{}

type Config struct {
	values map[string]any
	mutex  sync.RWMutex
}

func Get(key string) any {
	config.mutex.RLock()
	val := config.values[key]
	config.mutex.RUnlock()
	return val
}

func SetString(key string, value string) {
	config.mutex.Lock()
	config.values[key] = value
	config.mutex.Unlock()
}

func SetSnowflake(key string, value discord.Snowflake) {
	SetString(key, value.String())
}

func GetString(key string) string {
	return Get(key).(string)
}

func GetSnowflake(key string) *discord.Snowflake {
	str := GetString(key)
	if str == "" {
		return nil
	}

	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		slog.Error("[Elaina] Failed to load config value as snowflake: \"" + key + "\"")
		return nil
	}
	s := discord.Snowflake(i)
	return &s
}

func InitializeConfig() (err error) {
	slog.Info("[Elaina] Loading config...")
	config.values = defaultValues()

	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("[Elaina] No config file found, using default values instead")
			return nil
		}
		return err
	}

	var loaded map[string]any
	if err = json.Unmarshal(file, &loaded); err != nil {
		return err
	}

	missing := false
	for k := range config.values {
		if v, exists := loaded[k]; exists {
			config.values[k] = v
		} else {
			missing = true
			slog.Warn("[Elaina] Config file is missing value for key: \"" + k + "\"")
		}
	}

	if missing {
		if err = SaveConfig(); err != nil { // Saving the config again
			return err
		}
		slog.Info("[Elaina] Generated missing config values in file")
	}

	slog.Info("[Elaina] Config loaded")
	return nil
}

func SaveConfig() error {
	config.mutex.RLock()
	enc, err := json.MarshalIndent(config.values, "", "	")
	if err != nil {
		return err
	}
	config.mutex.RUnlock()
	return os.WriteFile(configPath, enc, 0660)
}
