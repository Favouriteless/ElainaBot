package discord

import (
	"bytes"
	"encoding/json"
)

// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
const (
	RespTypeChannelMessage         = 4
	RespTypeDeferredChannelMessage = 5
	RespTypeDeferredUpdateMessage  = 6
	RespTypeUpdateMessage          = 7
	RespTypeAutocomplete           = 8
)

func SendInteractionResponse(response InteractionResponse, id Snowflake, token string) error {
	encResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}
	resp, err := Post(Url("interactions", id.String(), token, "callback"), bytes.NewReader(encResponse))
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func SendInteractionMessageResponse(message Message, id Snowflake, token string) error {
	return SendInteractionResponse(InteractionResponse{Type: RespTypeChannelMessage, Data: message}, id, token)
}

func EditInteractionResponse(content string, token string) error {
	body, err := json.Marshal(struct {
		Content string `json:"content"`
	}{content})
	if err != nil {
		return err
	}

	resp, err := Patch(Url("webhooks", application.id, token, "messages", "@original"), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type InteractionResponse struct {
	Type int         `json:"type"`
	Data interface{} `json:"data"`
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
