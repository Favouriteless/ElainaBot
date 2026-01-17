package discord

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const createInteractionResponseUrl = apiUrl + "/interactions/%s/%s/callback"

// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type

const RespTypeChannelMessage = 4
const RespTypeDeferredChannelMessage = 5
const RespTypeDeferredUpdateMessage = 6
const RespTypeUpdateMessage = 7
const RespTypeAutocomplete = 8

func (client *Client) SendInteractionResponse(id Snowflake, token string, response InteractionResponse) (resp *http.Response, err error) {
	encResponse, err := json.Marshal(response)
	if err != nil {
		return
	}
	return client.Post(fmt.Sprintf(createInteractionResponseUrl, id, token), encResponse, 3)
}

type InteractionResponse struct {
	Type int         `json:"type"`
	Data interface{} `json:"data"`
}

// InteractionResponseType is a type safe factory for representing https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
type InteractionResponseType[T any] struct {
	id int
}

// MessageResponse represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-messages
type MessageResponse struct {
	Tts             bool            `json:"tts,omitempty"`
	Content         string          `json:"content,omitempty"`
	AllowedMentions AllowedMentions `json:"allowed_mentions,omitempty"`
	Flags           int             `json:"flags,omitempty"` // If using deferred channel message, only valid flag is ephemeral
	Attachments     []Attachment    `json:"attachments,omitempty"`
	Poll            *Poll           `json:"poll,omitempty"`
	// TODO: Embed, attachments & components support
}

// AutocompleteResponse represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-autocomplete
type AutocompleteResponse struct {
	Choices []CommandOptionChoice `json:"choices"` // Max 25 length
}

var defaultErrResponse = InteractionResponse{
	Type: RespTypeChannelMessage,
	Data: Message{Content: "An error occurred executing this command :(", Flags: MsgFlagEphemeral},
}

func (client *Client) handleInteractionCommands(payload InteractionCreatePayload) { // Built-in event handler for dispatching application Commands
	if payload.Type != 2 { // https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-data
		return
	}

	var c ApplicationCommandData
	if err := json.Unmarshal(*payload.Data, &c); err != nil {
		slog.Error("Failed to parse application command data: " + err.Error())
		return
	}

	for _, command := range client.Commands {
		if c.Name == command.Name {
			slog.Info("Dispatching application command: " + c.Name)

			if err := command.Handler(c, payload.Id, payload.Token); err != nil {
				slog.Error("Error executing application command: ", slog.String("command", c.Name), slog.String("error", err.Error()))
				_, _ = client.SendInteractionResponse(payload.Id, payload.Token, defaultErrResponse)
			}
			return
		}
	}
	slog.Warn("Application command was dispatched but no handler was found: " + c.Name)
}
