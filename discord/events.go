package discord

const ( // Payload Opcodes as specified by https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-opcodes
	opDispatch          = 0  // Receive
	opHeartbeat         = 1  // Bidirectional
	opIdentify          = 2  // Send
	opPresenceUpdate    = 3  // Send
	opVoiceUpdate       = 4  // Send
	opResume            = 6  // Send
	opReconnect         = 7  // Receive
	opRequestMembers    = 8  // Send
	opInvalidSession    = 9  // Receive
	opHello             = 10 // Receive
	opHeartbeatAck      = 11 // Receive
	opRequestSoundboard = 31 // Send
)

// GatewayEvent represents any dispatch (opcode 0) event received or sent via gateway.
type GatewayEvent interface {
	Name() string     // String name of the event this type refers to https://discord.com/developers/docs/events/gateway-events#payload-structure
	New() interface{} // Returns a new instance of the event type, this will be passed to json.Unmarshal and later handled.
}

// GatewayEventHandler represents a handler for a GatewayEvent. Events may have multiple handlers, which are run in
// sequence, but only one instance of the event type will be created for each event received.
type GatewayEventHandler interface {
	Name() string                // String name of the event this type refers to https://discord.com/developers/docs/events/gateway-events#payload-structure
	Handle(*Client, interface{}) // Handler function for the event type.
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

// Resume is a non-standard event, it doesn't have a type. Opcode 6 instead
type Resume struct {
	Token       string `json:"token"`
	SessionId   string `json:"session_id"`
	SequenceNum int    `json:"seq"`
}

// Ready is sent by discord after the client has successfully identified itself and is ready to receive events. Ready
// also has a built-in handler for fetching the gateway resume details.
type Ready struct {
	ApiVersion       int         `json:"v"`
	User             User        `json:"user"`
	Guilds           []Guild     `json:"guilds"`
	SessionId        string      `json:"session_id"`
	ResumeGatewayUrl string      `json:"resume_gateway_url"`
	Application      Application `json:"application"`
}

func (r Ready) Name() string {
	return "READY"
}

func (r Ready) New() interface{} {
	return Ready{}
}
