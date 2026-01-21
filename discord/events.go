package discord

import (
	"encoding/json"
	"log/slog"
)

var Events = &EventDispatcher{
	Ready:             Event[ReadyPayload]{Name: "READY", builtin: []EventHandler[ReadyPayload]{readyEvent}},
	CreateMessage:     Event[CreateMessagePayload]{Name: "MESSAGE_CREATE"},
	UpdateMessage:     Event[UpdateMessagePayload]{Name: "MESSAGE_UPDATE", builtin: []EventHandler[UpdateMessagePayload]{updateMessageEvent}},
	DeleteMessage:     Event[DeleteMessagePayload]{Name: "MESSAGE_DELETE", builtin: []EventHandler[DeleteMessagePayload]{deleteMessageEvent}},
	BulkDeleteMessage: Event[BulkDeleteMessagePayload]{Name: "MESSAGE_DELETE_BULK"},
	ReactionAdd:       Event[ReactionAddPayload]{Name: "MESSAGE_REACTION_ADD"},
	ReactionRemove:    Event[ReactionRemovePayload]{Name: "MESSAGE_REACTION_REMOVE"},
	InteractionCreate: Event[InteractionCreatePayload]{Name: "INTERACTION_CREATE", builtin: []EventHandler[InteractionCreatePayload]{interactionCreateEvent}},
	UpdateChannel:     Event[UpdateChannelPayload]{Name: "CHANNEL_UPDATE", builtin: []EventHandler[UpdateChannelPayload]{updateChannelEvent}},
	DeleteChannel:     Event[DeleteChannelPayload]{Name: "CHANNEL_UPDATE", builtin: []EventHandler[DeleteChannelPayload]{deleteChannelEvent}},
	UpdateRole:        Event[UpdateRolePayload]{Name: "GUILD_ROLE_UPDATE", builtin: []EventHandler[UpdateRolePayload]{updateRoleEvent}},
	DeleteRole:        Event[DeleteRolePayload]{Name: "GUILD_ROLE_UPDATE", builtin: []EventHandler[DeleteRolePayload]{deleteRoleEvent}},
}

// Event represents a deserialization and handler dispatcher for a type of Event. Built-in handlers will
// always run LAST, as they may modify a ResourceCache.
type Event[T any] struct {
	Name     string
	handlers []EventHandler[T]
	builtin  []EventHandler[T]
}

// EventHandler is called when an event it is registered to fires. Events are called in order.
type EventHandler[T any] = func(T)

// Register an event handler to be run when an event of this type is received by the gatewayConnection. Multiple handlers can be
// registered for a single type.
func (event *Event[T]) Register(handler ...EventHandler[T]) {
	event.handlers = append(event.handlers, handler...)
}

type EventDispatcher struct {
	Ready             Event[ReadyPayload]
	CreateMessage     Event[CreateMessagePayload]
	UpdateMessage     Event[UpdateMessagePayload]
	DeleteMessage     Event[DeleteMessagePayload]
	BulkDeleteMessage Event[BulkDeleteMessagePayload]
	ReactionAdd       Event[ReactionAddPayload]
	ReactionRemove    Event[ReactionRemovePayload]
	InteractionCreate Event[InteractionCreatePayload]
	UpdateChannel     Event[UpdateChannelPayload]
	DeleteChannel     Event[DeleteChannelPayload]
	UpdateRole        Event[UpdateRolePayload]
	DeleteRole        Event[DeleteRolePayload]
}

// dispatch decodes the given json-encoded []byte and dispatches it as an event
func (event *Event[T]) dispatch(raw []byte) {
	if len(event.handlers) == 0 && len(event.builtin) == 0 {
		return
	}
	var data T
	if err := json.Unmarshal(raw, &data); err != nil {
		slog.Error("Failed to parse gateway event: " + err.Error())
		return
	}
	for _, handler := range event.handlers {
		handler(data)
	}
	for _, handler := range event.builtin {
		handler(data)
	}
}

// Not really a fan of how this is implemented, but I couldn't figure out how to maintain type safety during event handler
// registration without doing this.
func dispatchEvent(name string, raw []byte) {
	switch name {
	case Events.Ready.Name:
		Events.Ready.dispatch(raw)
	case Events.CreateMessage.Name:
		Events.CreateMessage.dispatch(raw)
	case Events.UpdateMessage.Name:
		Events.UpdateMessage.dispatch(raw)
	case Events.DeleteMessage.Name:
		Events.DeleteMessage.dispatch(raw)
	case Events.BulkDeleteMessage.Name:
		Events.BulkDeleteMessage.dispatch(raw)
	case Events.ReactionAdd.Name:
		Events.ReactionAdd.dispatch(raw)
	case Events.ReactionRemove.Name:
		Events.ReactionRemove.dispatch(raw)
	case Events.InteractionCreate.Name:
		Events.InteractionCreate.dispatch(raw)
	case Events.UpdateChannel.Name:
		Events.UpdateChannel.dispatch(raw)
	case Events.DeleteChannel.Name:
		Events.DeleteChannel.dispatch(raw)
	case Events.UpdateRole.Name:
		Events.UpdateRole.dispatch(raw)
	case Events.DeleteRole.Name:
		Events.DeleteRole.dispatch(raw)
	}
}
