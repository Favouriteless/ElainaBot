package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
)

const (
	HelloEmoji        = "hello_emoji"
	DefaultHelloEmoji = "default_hello_emoji"
)

func defaultValues() map[string]any {
	return map[string]any{
		"hello_emoji":         "elainastare:1462288926663512274",
		"default_hello_emoji": "elainastare:1462288926663512274",
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

func Set(key string, value any) {
	config.mutex.Lock()
	config.values[key] = value
	config.mutex.Unlock()
}

func GetString(key string) string {
	return Get(key).(string)
}

func InitializeConfig() (err error) {
	slog.Info("Loading config...")
	config.values = defaultValues()

	file, err := os.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No config file found, using default values instead")
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
			slog.Warn("Config file is missing value for key: \"" + k + "\"")
		}
	}

	if missing {
		if err = SaveConfig(); err != nil { // Saving the config again
			return err
		}
		slog.Info("Generated missing config values in file")
	}

	slog.Info("Config loaded")
	return nil
}

func SaveConfig() error {
	config.mutex.RLock()
	enc, err := json.MarshalIndent(config.values, "", "	")
	if err != nil {
		return err
	}
	config.mutex.RUnlock()
	return os.WriteFile("config.json", enc, 0660)
}
