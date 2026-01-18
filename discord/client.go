package discord

import (
	"log/slog"
	"net/http"
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

	Gateway  Gateway // Gateway connection information. Most clients should not directly interact with this.
	Events   EventDispatcher
	Commands []*ApplicationCommand
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient(name string, id string, secret string, token string, intents int) (*Client, error) {
	client := defaultClient(name, id, secret, token, intents)
	if err := client.initialise(); err != nil {
		return nil, err
	}
	return client, nil
}

// initialise is responsible for any post-instantiation setup required for clients
func (client *Client) initialise() error {

	return nil
}

func defaultClient(name string, id string, secret string, token string, intents int) *Client {
	client := Client{
		Name:    name,
		Http:    http.Client{Timeout: time.Second * 5},
		Id:      id,
		Secret:  secret,
		Token:   token,
		Gateway: defaultGateway(intents),
		Events:  defaultEvents(),
	}
	client.Events.InteractionCreate.Register(handleInteractionCommands) // Built-in event handler for dispatching interaction commands
	client.Events.Ready.Register(handleReady)                           // Built-in event handler for updating the gateway resume URL and session ID
	return &client
}

func handleReady(payload ReadyPayload, client *Client) {
	client.Gateway.url = payload.ResumeGatewayUrl
	client.Gateway.sessionId = payload.SessionId
	slog.Info("Gateway connection established")
}
