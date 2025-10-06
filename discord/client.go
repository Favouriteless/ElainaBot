package discord

import (
	"encoding/json"
	"net/http"
	"os"
)

const apiVersion = "10"
const apiEncoding = "json"
const apiUrl = "https://discord.com/api/v" + apiVersion

// Client represents the auth details of the discord client, all methods interfacing with Discord API will require a
// secrets to authenticate with.
type Client struct {
	Name   string      // Name of the discord bot
	Http   http.Client // HTTP client used for interacting with Discord's REST API
	Id     string
	Secret string
	Token  string

	gatewayUrl        string      // URL for connecting to Discord's Gateway API
	heartbeatInterval int         // Interval, in milliseconds, between sending heartbeat payloads to gateway
	gateway           chan []byte // Queue for writing to gateway. Any []byte written in this channel will automatically be sent to gateway as text
	sequence          *int        // The last sequence number the client received from gateway
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient(name string) (*Client, error) {
	client, err := loadClient()
	if err != nil {
		return nil, err
	}

	client.Name = name
	client.Http = http.Client{}
	if err = client.connectGateway(); err != nil {
		return nil, err
	}
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

// SendAuthorised signs the given HTTP request with the bot's auth header and sends it
func (client *Client) SendAuthorised(req *http.Request) (resp *http.Response, err error) {
	client.setAuthorization(&req.Header)
	return client.Http.Do(req)
}

func (client *Client) setAuthorization(header *http.Header) *http.Header {
	header.Set("User-Agent", "DiscordBot (https://maven.favouriteless.net, 2.0.0)")
	header.Set("Authorization", "Bot "+client.Token)
	return header
}

// Get sends an HTTP get request to the given URL signed with the bot's authorization token
func (client *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return client.SendAuthorised(req)
}
