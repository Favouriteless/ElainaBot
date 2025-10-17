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
	Name   string      // Name of the discord bot
	Http   http.Client // HTTP client used for interacting with Discord's REST API
	Id     string      // Client ID
	Secret string      // Client Secret
	Token  string      // Bot token

	Gateway Gateway // Gateway connection information. Most clients should not directly interact with this.
	Events  EventDispatcher
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient(name string, intents int) (*Client, error) {
	client, err := loadClient(name, intents)
	if err != nil {
		return nil, err
	}
	if err = client.initialise(); err != nil {
		return nil, err
	}
	return client, nil
}

// initialise is responsible for any post-instantiation setup required for clients
func (client *Client) initialise() error {
	client.Events.Ready.Register(func(payload ReadyPayload) {
		client.Gateway.url = payload.ResumeGatewayUrl
		client.Gateway.sessionId = payload.SessionId
	})
	return nil
}

// loadClient initialises a default Client instance and handles loading bot details from json such as client id, client
// secret and bot token
func loadClient(name string, intents int) (*Client, error) {
	contents, err := os.ReadFile("data/secrets.json")
	if err != nil {
		return nil, err
	}

	client := Client{
		Name: name,
		Http: http.Client{Timeout: time.Second * 10},
		Gateway: Gateway{
			sendBuffer: make(chan []byte, 10), // Arbitrary capacity to prevent blocking.
			intents:    intents,
		},
		Events: defaultEvents(),
	}

	if err = json.Unmarshal(contents, &client); err != nil {
		return nil, err
	}
	return &client, err
}
