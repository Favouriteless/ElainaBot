package discord

import (
	"encoding/json"
	"log/slog"
)

// GatewayEventType represents a deserialisation and handler dispatcher for a type of GatewayEvent
type GatewayEventType[T any] struct {
	Name     string
	handlers []func(T, *Client)
}

// Register an event handler to be run when an event of this type is received by the gateway. Multiple handlers can be
// registered for a single type.
func (event *GatewayEventType[T]) Register(handler ...func(T, *Client)) {
	event.handlers = append(event.handlers, handler...)
}

// dispatch decodes the given json-encoded []byte and dispatches it as an event
func (event *GatewayEventType[T]) dispatch(raw []byte, client *Client) {
	if len(event.handlers) == 0 {
		return
	}
	var data T
	if err := json.Unmarshal(raw, &data); err != nil {
		slog.Error("Failed to parse gateway event: " + err.Error())
		return
	}
	for _, handler := range event.handlers {
		handler(data, client)
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
		Ready:             GatewayEventType[ReadyPayload]{Name: "READY"},
		CreateMessage:     GatewayEventType[CreateMessagePayload]{Name: "MESSAGE_CREATE"},
		UpdateMessage:     GatewayEventType[UpdateMessagePayload]{Name: "MESSAGE_UPDATE"},
		DeleteMessage:     GatewayEventType[DeleteMessagePayload]{Name: "MESSAGE_DELETE"},
		BulkDeleteMessage: GatewayEventType[BulkDeleteMessagePayload]{Name: "MESSAGE_DELETE_BULK"},
		ReactionAdd:       GatewayEventType[ReactionAddPayload]{Name: "MESSAGE_REACTION_ADD"},
		ReactionRemove:    GatewayEventType[ReactionRemovePayload]{Name: "MESSAGE_REACTION_REMOVE"},
		InteractionCreate: GatewayEventType[InteractionCreatePayload]{Name: "INTERACTION_CREATE"},
	}
}

// Not really a fan of how this is implemented, but I couldn't figure out how to maintain type safety during event handler
// registration without doing this.
func (client *Client) dispatchEvent(name string, raw []byte) {
	d := client.Events
	switch name {
	case d.Ready.Name:
		d.Ready.dispatch(raw, client)
	case d.CreateMessage.Name:
		d.CreateMessage.dispatch(raw, client)
	case d.UpdateMessage.Name:
		d.UpdateMessage.dispatch(raw, client)
	case d.DeleteMessage.Name:
		d.DeleteMessage.dispatch(raw, client)
	case d.BulkDeleteMessage.Name:
		d.BulkDeleteMessage.dispatch(raw, client)
	case d.ReactionAdd.Name:
		d.ReactionAdd.dispatch(raw, client)
	case d.ReactionRemove.Name:
		d.ReactionRemove.dispatch(raw, client)
	case d.InteractionCreate.Name:
		d.InteractionCreate.dispatch(raw, client)
	}
}
