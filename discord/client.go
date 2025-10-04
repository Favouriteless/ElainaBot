package discord

import (
	"encoding/json"
	"net/http"
	"os"
)

const apiUrl = "https://discord.com/api/v10"

// Client represents the auth details of the discord client, all methods interfacing with Discord API will require a
// secrets to authenticate with.
type Client struct {
	Http       http.Client
	Id         string
	Secret     string
	Token      string
	GatewayUrl string
}

// CreateClient creates and initialises a discord client and its gateway connection
func CreateClient() (*Client, error) {
	client, err := loadClient()
	if err != nil {
		return nil, err
	}
	client.Http = http.Client{}
	if err = client.connect(); err != nil {
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
	req.Header.Set("User-Agent", "DiscordBot (https://maven.favouriteless.net, 2.0.0")
	req.Header.Set("Authorization", "Bot "+client.Token)
	return client.Http.Do(req)
}

// Get sends an HTTP get request to the given URL signed with the bot's authorization token
func (client *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return client.SendAuthorised(req)
}
