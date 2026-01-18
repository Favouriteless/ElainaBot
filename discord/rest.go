package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

func (client *Client) DeployCommand(command *ApplicationCommand, attempts int) (*http.Response, error) {
	enc, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(url("applications", client.Id, "commands"), bytes.NewReader(enc), attempts)
	return resp, err
}

func (client *Client) DeleteCommand(command Snowflake) (*http.Response, error) {
	resp, err := client.Delete(url("applications", client.Id, "commands", command), 3)
	return resp, err
}

func (client *Client) DeployAllCommands() {
	slog.Info("Deploying all application commands...")
	for _, com := range client.Commands {
		func() {
			resp, err := client.DeployCommand(com, 3)
			if err != nil {
				slog.Error("Error registering command: ", slog.String("command", com.Name), slog.String("error", err.Error()))
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case 200:
				slog.Info("Command updated successfully: " + com.Name)
			case 201:
				slog.Info("Command added successfully: " + com.Name)
			default:
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err) // Should never be hit
				}
				slog.Error("Command could not be created: ", slog.String("command", com.Name), slog.String("status_code", resp.Status), slog.String("body", string(body)))
			}
		}()
	}
}

func (client *Client) CreateMessage(channel Snowflake, content string, tts bool) (*Message, error) {
	body, err := json.Marshal(struct {
		Content string `json:"content"`
		Tts     bool   `json:"tts"`
	}{content, tts})
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(url("channels", channel, "messages"), bytes.NewReader(body), 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("message was not created: " + strconv.Itoa(resp.StatusCode))
	}

	var message Message
	if err = json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (client *Client) CreateReaction(channel Snowflake, message Snowflake, emoji Snowflake) error {
	resp, err := client.Put(url("channels", channel, "messages", message, "reactions", emoji, "@me"), nil, 3)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("reaction was not created: %s: %s", resp.Status, string(body)))
	}
	return nil
}
