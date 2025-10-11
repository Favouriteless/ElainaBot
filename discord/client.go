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

const ( // All gateway intent bits expressed as constants. These can be ORed
	IGuilds                 = 1 << 0
	IGuildMembers           = 1 << 1
	IGuildModeration        = 1 << 2
	IGuildExpressions       = 1 << 3
	IGuildIntegrations      = 1 << 4
	IGuildWebhooks          = 1 << 5
	IGuildInvites           = 1 << 6
	IGuildVoiceStates       = 1 << 7
	IGuildPresences         = 1 << 8
	IGuildMessages          = 1 << 9
	IGuildMessageReactions  = 1 << 10
	IGuildMessageTyping     = 1 << 11
	IDirectMessages         = 1 << 12
	IDirectMessageReactions = 1 << 13
	IDirectMessageTyping    = 1 << 14
	IMessageContent         = 1 << 15
	IGuildScheduledEvents   = 1 << 16
	IAutoModConfig          = 1 << 20
	IAutoModExec            = 1 << 21
	IGuildMessagePolls      = 1 << 24
	IDirectMessagePolls     = 1 << 25
)

// Client represents the auth and session details of the discord client, all methods interfacing with Discord API will
// require a client.
type Client struct {
	Name   string       // Name of the discord bot
	Http   *http.Client // HTTP client used for interacting with Discord's REST API
	Id     string       // Client ID
	Secret string       // Client Secret
	Token  string       // Bot token

	Gateway *Gateway // Gateway connection information. Most clients should not directly interact with this.
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
		Http: &http.Client{Timeout: time.Second * 10},
		Gateway: &Gateway{
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
