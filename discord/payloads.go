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
// also has a built-in handler for fetching the gateway resume details
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

// CreateMessagePayload is sent by discord when a message is created in one of the guilds the application is being used in
type CreateMessagePayload struct {
	Message
	GuildId *Snowflake `json:"guild_id"` // Optional
	// Member
	Mentions []User
}

func (p CreateMessagePayload) Name() string {
	return "MESSAGE_CREATE"
}
