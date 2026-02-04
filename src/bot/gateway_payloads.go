package main

import . "elaina-common"

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

// helloPayload is a non-standard event, it doesn't have a type. Opcode 1 instead
type helloPayload struct {
	HeartbeatInterval int64 `json:"heartbeat_interval"`
}

// identifyPayload is a non-standard event, it doesn't have a type. Opcode 2 instead
type identifyPayload struct {
	Token      string               `json:"token"`
	Properties connectionProperties `json:"properties"`
	Intents    int                  `json:"intents"`
}

// resumePayload is a non-standard event, it doesn't have a type. Opcode 6 instead
type resumePayload struct {
	Token       string `json:"token"`
	SessionId   string `json:"session_id"`
	SequenceNum int32  `json:"seq"`
}

// readyPayload is sent by discord after the client has successfully identified itself and is ready to receive events.
// https://discord.com/developers/docs/events/gateway-events#ready
type readyPayload struct {
	ApiVersion       int         `json:"v"`
	User             User        `json:"user"`
	Guilds           []Guild     `json:"guilds"`
	SessionId        string      `json:"session_id"`
	ResumeGatewayUrl string      `json:"resume_gateway_url"`
	Application      Application `json:"application"`
}
