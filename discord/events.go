package discord

const (
	opDispatch          = 0
	opHeartbeat         = 1
	opIdentify          = 2
	opPresenceUpdate    = 3
	opVoiceUpdate       = 4
	opResume            = 6
	opRequestMembers    = 8
	opInvalidSession    = 9
	opHello             = 10
	opHeartbeatAck      = 11
	opRequestSoundboard = 31
)

type GatewayEvent interface {
	Type() string
}

// Hello is a non-standard event, it doesn't have a type. Opcode 1 instead
type Hello struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// Identify is a non-standard event, it doesn't have a type. Opcode 2 instead
type Identify struct {
	Token      string               `json:"token"`
	Properties ConnectionProperties `json:"properties"`
	Intents    int                  `json:"intents"`
}

// Ready is sent by discord after the client has successfully identified itself and is ready to receive events
type Ready struct {
	ApiVersion       int         `json:"v"`
	User             User        `json:"user"`
	Guilds           []Guild     `json:"guilds"`
	SessionId        string      `json:"session_id"`
	ResumeGatewayUrl string      `json:"resume_gateway_url"`
	Application      Application `json:"application"`
}
