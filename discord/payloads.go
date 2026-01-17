package discord

// HelloPayload is a non-standard event, it doesn't have a type. Opcode 1 instead
type HelloPayload struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyPayload is a non-standard event, it doesn't have a type. Opcode 2 instead
type IdentifyPayload struct {
	Token      string               `json:"token"`
	Properties ConnectionProperties `json:"properties"`
	Intents    int                  `json:"intents"`
}

// ResumePayload is a non-standard event, it doesn't have a type. Opcode 6 instead
type ResumePayload struct {
	Token       string `json:"token"`
	SessionId   string `json:"session_id"`
	SequenceNum int    `json:"seq"`
}

// ReadyPayload is sent by discord after the client has successfully identified itself and is ready to receive events. It
// also has a built-in handler for fetching the gateway resume details.
// https://discord.com/developers/docs/events/gateway-events#ready
type ReadyPayload struct {
	ApiVersion       int         `json:"v"`
	User             User        `json:"user"`
	Guilds           []Guild     `json:"guilds"`
	SessionId        string      `json:"session_id"`
	ResumeGatewayUrl string      `json:"resume_gateway_url"`
	Application      Application `json:"application"`
}

func (p ReadyPayload) Name() string {
	return "RESUME"
}

// CreateMessagePayload is sent by discord when a message is created.
// https://discord.com/developers/docs/events/gateway-events#message-create
type CreateMessagePayload struct {
	Message
	GuildId  *Snowflake   `json:"guild_id"` // Optional
	Member   *GuildMember `json:"member"`   // Optional
	Mentions []User
}

func (p CreateMessagePayload) Name() string {
	return "MESSAGE_CREATE"
}

// UpdateMessagePayload is sent by discord when a message is edited/updated.
// https://discord.com/developers/docs/events/gateway-events#message-update
type UpdateMessagePayload struct {
	Message
	GuildId  *Snowflake   `json:"guild_id"` // Optional
	Member   *GuildMember `json:"member"`   // Optional
	Mentions []User
}

// DeleteMessagePayload is sent by discord when a single message is deleted.
// https://discord.com/developers/docs/events/gateway-events#message-delete
type DeleteMessagePayload struct {
	Id        Snowflake  `json:"id"`
	ChannelId Snowflake  `json:"channel_id"`
	GuildId   *Snowflake `json:"guild_id"` // Optional
}

// BulkDeleteMessagePayload is sent by discord when a multiple messages are deleted.
// https://discord.com/developers/docs/events/gateway-events#message-delete-bulk
type BulkDeleteMessagePayload struct {
	Ids       []Snowflake `json:"ids"`
	ChannelId Snowflake   `json:"channel_id"`
	GuildId   *Snowflake  `json:"guild_id"` // Optional
}

// ReactionAddPayload is sent by discord when a user adds a reaction to a message.
// https://discord.com/developers/docs/events/gateway-events#message-reaction-add
type ReactionAddPayload struct {
	UserId          Snowflake    `json:"user_id"`
	ChannelId       Snowflake    `json:"channel_id"`
	MessageId       Snowflake    `json:"message_id"`
	GuildId         *Snowflake   `json:"guild_id"` // Optional
	Member          *GuildMember `json:"member"`   // Optional
	Emoji           Emoji        `json:"emoji"`
	MessageAuthorId *Snowflake   `json:"message_author_id"` // Optional
	Burst           bool         `json:"burst"`
	BurstColors     []string     `json:"burst_colors"` // Optional
	Type            int          `json:"type"`
}

// ReactionRemovePayload is sent by discord when a user removes a reaction from a message.
// https://discord.com/developers/docs/events/gateway-events#message-reaction-remove
type ReactionRemovePayload struct {
	UserId    Snowflake  `json:"user_id"`
	ChannelId Snowflake  `json:"channel_id"`
	MessageId Snowflake  `json:"message_id"`
	GuildId   *Snowflake `json:"guild_id"` // Optional
	Emoji     Emoji      `json:"emoji"`
	Burst     bool       `json:"burst"`
	Type      int        `json:"type"`
}

// InteractionCreatePayload is sent by discord when a user creates an interaction (e.g. via a slash command)
type InteractionCreatePayload = Interaction
