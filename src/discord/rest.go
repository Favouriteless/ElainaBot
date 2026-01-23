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

func getCacheable[T any](cache *ResourceCache[T], id Snowflake, urlParts ...string) (*T, error) {
	if val := cache.Get(id); val != nil {
		return val, nil
	}

	resp, err := Get(Url(urlParts...))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var val T
	if err = json.NewDecoder(resp.Body).Decode(&val); err != nil {
		return nil, err
	}
	cache.Add(id, val)
	return &val, nil
}

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

func (channel Channel) CreateMessage(content string, tts bool) (*Message, error) {
	return CreateMessage(channel.Id, content, tts)
}

func GetChannel(id Snowflake) (*Channel, error) {
	return getCacheable(ChannelCache, id, "channels", id.String())
}

func GetRole(guildId Snowflake, roleId Snowflake) (*Role, error) {
	return getCacheable(RoleCache, roleId, "guilds", guildId.String(), "roles", roleId.String())
}

func (guild Guild) GetRole(id Snowflake) (*Role, error) {
	return GetRole(guild.Id, id)
}

func GetGuildMember(guild Snowflake, guildMemberId Snowflake) (*GuildMember, error) {
	return getCacheable(GuildMemberCache, guildMemberId, "guilds", guild.String(), "members", guildMemberId.String())
}

func (guild Guild) GetGuildMember(id Snowflake) (*GuildMember, error) {
	return GetGuildMember(guild.Id, id)
}

func GetGuild(id Snowflake) (*Guild, error) {
	return getCacheable(GuildCache, id, "guilds", id.String())
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

func (channel Channel) CreateReaction(messageId Snowflake, emoji string) error {
	return CreateReaction(channel.Id, messageId, emoji)
}

func (message Message) CreateReaction(emoji string) error {
	return CreateReaction(message.ChannelId, message.Id, emoji)
}

func CreateBan(guildId Snowflake, userId Snowflake, deleteSeconds int) error {
	enc, err := json.Marshal(struct {
		Seconds int `json:"delete_message_seconds"`
	}{deleteSeconds})
	if err != nil {
		return err
	}

	resp, err := Put(Url("guilds", guildId.String(), "bans", userId.String()), enc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (guild Guild) CreateBan(userId Snowflake, deleteSeconds int) error {
	return CreateBan(guild.Id, userId, deleteSeconds)
}

func (user User) CreateBan(guildId Snowflake, deleteSeconds int) error {
	return CreateBan(guildId, user.Id, deleteSeconds)
}
