package discord

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

const apiVersion = "10"
const apiEncoding = "json"
const apiUrl = "https://discord.com/api/v" + apiVersion

// Client represents the auth and session details of the discord client, all methods interfacing with Discord API will
// require a client.
type Client struct {
	Name   string       // Name of the discord bot
	Http   *http.Client // HTTP client used for interacting with Discord's REST API
	Id     string       // Client ID
	Secret string       // Client Secret
	Token  string       // Bot token

	Gateway *Gateway // Gateway connection information. Most clients should not directly interact with this.
	// TODO: It would probably be better for the client to track the gateway connection, reader/writer stops, etc. itself so it can close or reconnect at will.
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient(name string, timeout time.Duration) (*Client, error) {
	client, err := loadClient()
	if err != nil {
		return nil, err
	}

	client.Name = name
	client.Http = &http.Client{Timeout: timeout}

	return client, nil
}

// loadClient handles loading bot details from json such as client id, client secret and bot token
func loadClient() (*Client, error) {
	contents, err := os.ReadFile("data/secrets.json")
	if err != nil {
		return nil, err
	}
	var client Client
	if err = json.Unmarshal(contents, &client); err != nil {
		return nil, err
	}
	return &client, err
}
