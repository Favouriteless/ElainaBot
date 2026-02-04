package main

import (
	. "elaina-common"
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

func defaultConfigValues() map[string]string {
	return map[string]string{
		HelloEmoji:        "elainastare:1462289034188689468",
		DefaultHelloEmoji: "elainastare:1462289034188689468",
	}
}

var config = &Config{}

type Config struct {
	values map[string]string
	mutex  sync.RWMutex
}

func getConfig(key string) string {
	config.mutex.RLock()
	val := config.values[key]
	config.mutex.RUnlock()
	return val
}

func setConfigString(key string, value string) {
	config.mutex.Lock()
	config.values[key] = value
	config.mutex.Unlock()
}

func setConfigSnowflake(key string, value Snowflake) {
	setConfigString(key, value.String())
}

func getConfigSnowflake(key string) *Snowflake {
	str := getConfig(key)
	if str == "" {
		return nil
	}

	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		slog.Error("[Elaina] Failed to load config value as snowflake: \"" + key + "\"")
		return nil
	}
	s := Snowflake(i)
	return &s
}

func initializeConfig() (err error) {
	slog.Info("[Elaina] Loading config...")
	config.values = defaultConfigValues()

	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("[Elaina] No config file found, using default values instead")
			return nil
		}
		return err
	}

	var loaded map[string]string
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
		if err = saveConfig(); err != nil { // Saving the config again
			return err
		}
		slog.Info("[Elaina] Generated missing config values in file")
	}

	slog.Info("[Elaina] Config loaded")
	return nil
}

func saveConfig() error {
	config.mutex.RLock()
	enc, err := json.MarshalIndent(config.values, "", "	")
	if err != nil {
		return err
	}
	config.mutex.RUnlock()
	return os.WriteFile(configPath, enc, 0660)
}
