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

	gateway struct {
		gatewayUrl       string      // URL for connecting to Discord's Gateway API
		resumeGatewayUrl string      // URL for resuming a gateway connection
		sessionId        string      // ID of gateway session, only applicable if resuming
		sendBuffer       chan []byte // Queue for writing to gateway. Any []byte written in this channel will automatically be sent to gateway as text
		sequence         *int        // The last sequence number the client received from gateway

		heartbeatAcknowledged bool // Set to false when the client sends a heartbeat. If discord doesn't acknowledge before the next heartbeat, we reconnect. No mutex needed as booleans don't tear and there's a large interval
	}
	// TODO: It would probably be better for the client to track the gateway connection, reader/writer stops, etc. itself so it can close or reconnect at will.
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient(name string, timeout time.Duration) (*Client, error) {
	client, err := loadClient()
	if err != nil {
		return nil, err
	}

	client.Name = name
	client.Http = http.Client{Timeout: timeout}

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
