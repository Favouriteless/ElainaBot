package discord

import (
	"encoding/json"
	"log"
)

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

// GatewayEvent represents a discord gateway event with a name associated with it (HELLO, IDENTIFY and RESUME are not named)
type GatewayEvent interface {
	Name() string // String name of the event this type refers to https://discord.com/developers/docs/events/gateway-events#payload-structure
}

// GatewayEventType represents a deserialisation and handler dispatcher for a type of GatewayEvent
type GatewayEventType[T GatewayEvent] struct {
	Default  T // Default instance of the event payload to deserialize into a copy of
	handlers []func(T)
}

func (event *GatewayEventType[T]) Name() string {
	return event.Default.Name()
}

// Register an event handler to be run when an event of this type is received by the gateway. Multiple handlers can be
// registered for a single type.
func (event *GatewayEventType[T]) Register(handler func(T)) {
	event.handlers = append(event.handlers, handler)
}

func (event *GatewayEventType[T]) dispatch(raw []byte) {
	data := event.Default // Creates a shallow copy
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Printf("Failed to dispatch gateway event: %s", err)
		return
	}
	for _, handler := range event.handlers {
		handler(data)
	}
}

type EventDispatcher struct {
	Ready         GatewayEventType[ReadyPayload]
	CreateMessage GatewayEventType[CreateMessagePayload]
}

func defaultEvents() EventDispatcher {
	return EventDispatcher{
		Ready:         GatewayEventType[ReadyPayload]{Default: ReadyPayload{}},
		CreateMessage: GatewayEventType[CreateMessagePayload]{Default: CreateMessagePayload{}},
	}
}

func (d *EventDispatcher) dispatchEvent(name string, raw []byte) {
	switch name { // TODO: Fill in remaining event handlers as needed
	case d.Ready.Name():
		d.Ready.dispatch(raw)
	case d.CreateMessage.Name():
		d.CreateMessage.dispatch(raw)
	}
}
