package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
)

func DeployCommand(command *ApplicationCommand) (*http.Response, error) {
	enc, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	resp, err := Post(Url("applications", application.id, "commands"), bytes.NewReader(enc))
	return resp, err
}

func DeleteCommand(command Snowflake) (*http.Response, error) {
	resp, err := Delete(Url("applications", application.id, "commands", command.String()))
	return resp, err
}

func DeployCommands(names ...string) {
	slog.Info("Deploying all application commands...")
	for _, com := range Commands {
		if slices.Contains(names, com.Name) {
			func() {
				resp, err := DeployCommand(com)
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
}

func CreateMessage(channel Snowflake, content string, tts bool) (*Message, error) {
	body, err := json.Marshal(struct {
		Content string `json:"content"`
		Tts     bool   `json:"tts"`
	}{content, tts})
	if err != nil {
		return nil, err
	}

	resp, err := Post(Url("channels", channel.String(), "messages"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("message was not created: %s", resp.Status)
	}

	var message Message
	if err = json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, err
	}
	return &message, nil
}

func GetChannel(id Snowflake) (*Channel, error) {
	if channel := ChannelCache.Get(id); channel != nil {
		return channel, nil
	}

	resp, err := Get(Url("channels", id.String()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var channel Channel
	if err = json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return nil, err
	}
	ChannelCache.Add(id, channel)
	return &channel, nil
}

func GetRole(guildId Snowflake, roleId Snowflake) (*Role, error) {
	if role := RoleCache.Get(roleId); role != nil {
		return role, nil
	}

	resp, err := Get(Url("guilds", guildId.String(), "roles", roleId.String()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var role Role
	if err = json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, err
	}
	RoleCache.Add(roleId, role)
	return &role, nil
}

func GetGuildMember(guild Snowflake, guildMemberId Snowflake) (*GuildMember, error) {
	resp, err := Get(Url("guilds", guild.String(), "members", guildMemberId.String()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var guildMember GuildMember
	if err = json.NewDecoder(resp.Body).Decode(&guildMember); err != nil {
		return nil, err
	}
	return &guildMember, nil
}

// CreateReaction creates a reaction to a message using the bot account. emoji must be either a Unicode emoji for
// built-in emojis or a string in the format "name:snowflake" for custom discord emojis.
func CreateReaction(channelId Snowflake, messageId Snowflake, emoji string) error {
	resp, err := Put(Url("channels", channelId.String(), "messages", messageId.String(), "reactions", emoji, "@me"), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("reaction was not created: %s: %s", resp.Status, string(body))
	}
	return nil
}
