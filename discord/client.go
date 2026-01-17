package discord

import (
	"encoding/json"
	"io"
	"log"
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

	Gateway  Gateway // Gateway connection information. Most clients should not directly interact with this.
	Events   EventDispatcher
	Commands []*ApplicationCommand
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
	client.Events.Ready.Register(func(payload ReadyPayload) { // Built-in event handler for updating the gateway resume URL and session ID
		client.Gateway.url = payload.ResumeGatewayUrl
		client.Gateway.sessionId = payload.SessionId
	})

	client.Events.InteractionCreate.Register(func(payload InteractionCreatePayload) { // Built-in event handler for dispatching application Commands
		if payload.Type == 2 { // https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-data
			var c ApplicationCommandData
			if err := json.Unmarshal(*payload.Data, &c); err != nil {
				log.Println("Error parsing command:", err)
				return
			}

			log.Println("Dispatching command: " + c.Name)
			for _, command := range client.Commands {
				if c.Name == command.Name && command.Handler(c) {
					break
				}
			}
		}
	})
	return nil
}

// loadClient initialises a default Client instance and handles loading bot details from JSON such as client id, client
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
			sendQueue: make(chan []byte, 10), // Arbitrary capacity to prevent blocking.
			intents:   intents,
		},
		Events: defaultEvents(),
	}

	if err = json.Unmarshal(contents, &client); err != nil {
		return nil, err
	}
	return &client, err
}

func (client *Client) DeployCommand(command *ApplicationCommand, attempts int) (*http.Response, error) {
	enc, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post(apiUrl+"/applications/"+client.Id+"/commands", enc, attempts)
	return resp, err
}

func (client *Client) DeleteCommand(command Snowflake) (*http.Response, error) {
	resp, err := client.Delete(apiUrl+"/applications/"+client.Id+"/commands/"+command, 3)
	return resp, err
}

func (client *Client) DeployAllCommands() {
	for _, com := range client.Commands {
		func() {
			resp, err := client.DeployCommand(com, 3)
			if err != nil {
				log.Printf("Error registering command \"%s\": %s", com.Name, err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case 200:
				log.Printf("Command \"%s\" already exists, it was updated", com.Name)
			case 201:
				log.Printf("Command \"%s\" added successfully", com.Name)
			default:
				log.Printf("Command \"%s\" could not be created: %s", com.Name, resp.Status)
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err) // Should never be hit
				}
				log.Println(string(body))
			}
		}()
	}
}
