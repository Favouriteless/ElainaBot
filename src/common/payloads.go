package common

import "time"

// CreateMessagePayload is sent by discord when a message is created.
// https://discord.com/developers/docs/events/gateway-events#message-create
type CreateMessagePayload struct {
	Message
	GuildId  Snowflake    `json:"guild_id"` // Optional
	Member   *GuildMember `json:"member"`   // Optional
	Mentions []User
}

// UpdateMessagePayload is sent by discord when a message is edited/updated.
// https://discord.com/developers/docs/events/gateway-events#message-update
type UpdateMessagePayload struct {
	Message
	GuildId  Snowflake    `json:"guild_id"` // Optional
	Member   *GuildMember `json:"member"`   // Optional
	Mentions []User
}

// DeleteMessagePayload is sent by discord when a single message is deleted.
// https://discord.com/developers/docs/events/gateway-events#message-delete
type DeleteMessagePayload struct {
	Id        Snowflake `json:"id"`
	ChannelId Snowflake `json:"channel_id"`
	GuildId   Snowflake `json:"guild_id"` // Optional
}

// BulkDeleteMessagePayload is sent by discord when a multiple messages are deleted.
// https://discord.com/developers/docs/events/gateway-events#message-delete-bulk
type BulkDeleteMessagePayload struct {
	Ids       []Snowflake `json:"ids"`
	ChannelId Snowflake   `json:"channel_id"`
	GuildId   Snowflake   `json:"guild_id"` // Optional
}

// ReactionAddPayload is sent by discord when a user adds a reaction to a message.
// https://discord.com/developers/docs/events/gateway-events#message-reaction-add
type ReactionAddPayload struct {
	UserId          Snowflake    `json:"user_id"`
	ChannelId       Snowflake    `json:"channel_id"`
	MessageId       Snowflake    `json:"message_id"`
	GuildId         Snowflake    `json:"guild_id"` // Optional
	Member          *GuildMember `json:"member"`   // Optional
	Emoji           Emoji        `json:"emoji"`
	MessageAuthorId Snowflake    `json:"message_author_id"` // Optional
	Burst           bool         `json:"burst"`
	BurstColors     []string     `json:"burst_colors"` // Optional
	Type            int          `json:"type"`
}

// ReactionRemovePayload is sent by discord when a user removes a reaction from a message.
// https://discord.com/developers/docs/events/gateway-events#message-reaction-remove
type ReactionRemovePayload struct {
	UserId    Snowflake `json:"user_id"`
	ChannelId Snowflake `json:"channel_id"`
	MessageId Snowflake `json:"message_id"`
	GuildId   Snowflake `json:"guild_id"` // Optional
	Emoji     Emoji     `json:"emoji"`
	Burst     bool      `json:"burst"`
	Type      int       `json:"type"`
}

// InteractionCreatePayload is sent by discord when a user creates an interaction (e.g. via a slash command)
type InteractionCreatePayload = Interaction

// UpdateChannelPayload is sent by discord when a guild channel is updated.
// https://discord.com/developers/docs/events/gateway-events#channel-update
type UpdateChannelPayload = Channel

// DeleteChannelPayload is sent by discord when a guild channel is deleted.
// https://discord.com/developers/docs/events/gateway-events#channel-delete
type DeleteChannelPayload = Channel

// UpdateRolePayload is sent by discord when a guild channel is updated.
// https://discord.com/developers/docs/events/gateway-events#guild-role-update
type UpdateRolePayload struct {
	GuildId Snowflake `json:"guild_id"`
	Role    Role      `json:"role"`
}

// DeleteRolePayload is sent by discord when a guild channel is deleted.
// https://discord.com/developers/docs/events/gateway-events#guild-role-delete
type DeleteRolePayload struct {
	GuildId Snowflake `json:"guild_id"`
	RoleId  Snowflake `json:"role_id"`
}

// UpdateGuildPayload is sent by discord when a guild is updated.
// https://discord.com/developers/docs/events/gateway-events#guild-update
type UpdateGuildPayload = Guild

// DeleteGuildPayload is sent by discord when a guild is created, becomes unavailable or the user leaves a guild.
// https://discord.com/developers/docs/events/gateway-events#guild-delete
type DeleteGuildPayload = UnavailableGuild

// ModifyGuildMemberPayload is sent to discord to update a GuildMember resource.
// https://discord.com/developers/docs/resources/guild#modify-guild-member
type ModifyGuildMemberPayload struct {
	Nick                       *Nullable[string]    `json:"nick,omitempty"`
	Roles                      []Snowflake          `json:"roles,omitzero"`
	Mute                       *bool                `json:"mute,omitempty"`
	Deaf                       *bool                `json:"deaf,omitempty"`
	ChannelId                  *Nullable[Snowflake] `json:"channel_id,omitempty"`
	CommunicationDisabledUntil *Nullable[time.Time] `json:"communication_disabled_until,omitempty"`
}
