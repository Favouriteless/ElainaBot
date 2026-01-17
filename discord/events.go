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

// GatewayEventType represents a deserialisation and handler dispatcher for a type of GatewayEvent
type GatewayEventType[T any] struct {
	Name     string
	handlers []func(T)
}

// Register an event handler to be run when an event of this type is received by the gateway. Multiple handlers can be
// registered for a single type.
func (event *GatewayEventType[T]) Register(handler func(T)) {
	event.handlers = append(event.handlers, handler)
}

// dispatch decodes the given json-encoded []byte and dispatches it as an event
func (event *GatewayEventType[T]) dispatch(raw []byte) {
	if len(event.handlers) == 0 {
		return
	}
	var data T
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Printf("Failed to dispatch gateway event: %s", err)
		return
	}
	for _, handler := range event.handlers {
		handler(data)
	}
}

type EventDispatcher struct {
	Ready             GatewayEventType[ReadyPayload]
	CreateMessage     GatewayEventType[CreateMessagePayload]
	UpdateMessage     GatewayEventType[UpdateMessagePayload]
	DeleteMessage     GatewayEventType[DeleteMessagePayload]
	BulkDeleteMessage GatewayEventType[BulkDeleteMessagePayload]
	ReactionAdd       GatewayEventType[ReactionAddPayload]
	ReactionRemove    GatewayEventType[ReactionRemovePayload]
	InteractionCreate GatewayEventType[InteractionCreatePayload]
}

func defaultEvents() EventDispatcher {
	return EventDispatcher{ // Do not use explicit names here, we want the compiler to complain if an event is missing
		GatewayEventType[ReadyPayload]{Name: "READY"},
		GatewayEventType[CreateMessagePayload]{Name: "MESSAGE_CREATE"},
		GatewayEventType[UpdateMessagePayload]{Name: "MESSAGE_UPDATE"},
		GatewayEventType[DeleteMessagePayload]{Name: "MESSAGE_DELETE"},
		GatewayEventType[BulkDeleteMessagePayload]{Name: "MESSAGE_DELETE_BULK"},
		GatewayEventType[ReactionAddPayload]{Name: "MESSAGE_REACTION_ADD"},
		GatewayEventType[ReactionRemovePayload]{Name: "MESSAGE_REACTION_REMOVE"},
		GatewayEventType[InteractionCreatePayload]{Name: "INTERACTION_CREATE"},
	}
}

// Not really a fan of how this is implemented, but I couldn't figure out how to maintain type safety during event handler
// registration without doing this.
func (d *EventDispatcher) dispatchEvent(name string, raw []byte) {
	switch name { // TODO: Fill in remaining event handlers as needed
	case d.Ready.Name:
		d.Ready.dispatch(raw)
	case d.CreateMessage.Name:
		d.CreateMessage.dispatch(raw)
	case d.UpdateMessage.Name:
		d.UpdateMessage.dispatch(raw)
	case d.DeleteMessage.Name:
		d.DeleteMessage.dispatch(raw)
	case d.BulkDeleteMessage.Name:
		d.BulkDeleteMessage.dispatch(raw)
	case d.ReactionAdd.Name:
		d.ReactionAdd.dispatch(raw)
	case d.ReactionRemove.Name:
		d.ReactionRemove.dispatch(raw)
	case d.InteractionCreate.Name:
		d.InteractionCreate.dispatch(raw)
	}
}
