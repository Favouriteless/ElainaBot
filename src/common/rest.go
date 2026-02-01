package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var RoleCache = CreateCache[Snowflake, Role](10)
var MessageCache = CreateCache[Snowflake, Message](50)
var ChannelCache = CreateCache[Snowflake, Channel](20)
var GuildCache = CreateCache[Snowflake, Guild](3)
var GuildMemberCache = CreateCache[Snowflake, GuildMember](10)

func getCacheable[K comparable, T any](cache *LRUCache[K, T], id K, urlParts ...string) (*T, error) {
	if val := cache.Get(id); val != nil {
		return val, nil
	}

	resp, err := Get(ApiUrl(urlParts...))
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

func CreateOrUpdateCommand(command *ApplicationCommand) (*http.Response, error) {
	enc, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	resp, err := Post(ApiUrl("applications", CommonSecrets.Id, "commands"), bytes.NewReader(enc))
	return resp, err
}

func DeleteCommand(command Snowflake) (*http.Response, error) {
	resp, err := Delete(ApiUrl("applications", CommonSecrets.Id, "commands", command.String()))
	return resp, err
}

func CreateMessage(channel Snowflake, content string, tts bool) (*Message, error) {
	body, err := json.Marshal(struct {
		Content string `json:"content"`
		Tts     bool   `json:"tts"`
	}{content, tts})
	if err != nil {
		return nil, err
	}

	resp, err := Post(ApiUrl("channels", channel.String(), "messages"), bytes.NewReader(body))
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

func DeleteMessage(channel Snowflake, message Snowflake) error {
	resp, err := Delete(ApiUrl("channels", channel.String(), "messages", message.String()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("message was not deleted: %s", resp.Status)
	}
	return nil
}

func CreateDM(recipient Snowflake) (*Channel, error) {
	body, err := json.Marshal(struct {
		Recipient Snowflake `json:"recipient_id"`
	}{recipient})
	if err != nil {
		return nil, err
	}

	resp, err := Post(ApiUrl("users", "@me", "channels"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("dm was not created: %s: %s", resp.Status, string(body))
	}

	var channel Channel
	if err = json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

func (user User) CreateDM() (*Channel, error) {
	return CreateDM(user.Id)
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
	resp, err := Put(ApiUrl("channels", channelId.String(), "messages", messageId.String(), "reactions", emoji, "@me"), nil)
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

	resp, err := Put(ApiUrl("guilds", guildId.String(), "bans", userId.String()), enc)
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

func DeleteBan(guildId Snowflake, userId Snowflake) error {
	resp, err := Delete(ApiUrl("guilds", guildId.String(), "bans", userId.String()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func KickUser(guildId Snowflake, userId Snowflake) error {
	resp, err := Delete(ApiUrl("guilds", guildId.String(), "members", userId.String()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func ModifyGuildMember(guildId Snowflake, userId Snowflake, payload ModifyGuildMemberPayload) error {
	enc, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := Patch(ApiUrl("guilds", guildId.String(), "members", userId.String()), enc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (guild Guild) KickUser(userId Snowflake) error {
	return KickUser(guild.Id, userId)
}

func GetUserMessages(guildId Snowflake, author Snowflake, start time.Time, end time.Time) ([]Message, error) {
	startSnowflake := TimeToSnowflake(start)
	endSnowflake := TimeToSnowflake(end) + 4194303

	url := ApiUrl("guilds", guildId.String(), "messages", "search")
	url += QueryParams(
		"author_id", author.String(),
		"include_nsfw", "true",
		"min_id", startSnowflake.String(),
		"max_id", endSnowflake.String())

	resp, err := Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := struct {
		AnalyticsId      string      `json:"analytics_id"`
		Messages         [][]Message `json:"messages"`
		TotalResults     int
		DocumentsIndexed int `json:"documents_indexed"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	messages := make([]Message, response.TotalResults)
	for _, m := range response.Messages {
		for _, message := range m {
			messages = append(messages, message)
		}
	}

	return messages, nil
}
