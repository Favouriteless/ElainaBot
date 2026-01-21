package discord

import (
	"encoding/json"
	"log/slog"
)

// readyEvent updates the gateway resuming URL and session ID
func readyEvent(payload ReadyPayload) {
	gatewayConnection.resumeUrl = payload.ResumeGatewayUrl
	gatewayConnection.sessionId = payload.SessionId
	slog.Info("Gateway connection established")
}

// interactionCreateEvent dispatches ApplicationCommands
func interactionCreateEvent(payload InteractionCreatePayload) { // Built-in event handler for dispatching application Commands
	if payload.Type != 2 { // https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-data
		return
	}

	var c ApplicationCommandData
	if err := json.Unmarshal(*payload.Data, &c); err != nil {
		slog.Error("Failed to parse application command data: " + err.Error())
		return
	}

	for _, command := range Commands {
		if c.Name == command.Name {
			slog.Info("Dispatching application command: " + c.Name)

			if err := command.Handler(c, payload.Id, payload.Token); err != nil {
				slog.Error("Error executing application command: ", slog.String("command", c.Name), slog.String("error", err.Error()))
				_ = SendInteractionResponse(InteractionResponse{
					Type: RespTypeChannelMessage,
					Data: Message{Content: "An error occurred executing this command: " + err.Error(), Flags: MsgFlagEphemeral},
				}, payload.Id, payload.Token)
			}
			return
		}
	}
	slog.Warn("Application command was dispatched but no handler was found: " + c.Name)
}

func updateChannelEvent(payload UpdateChannelPayload) {
	ChannelCache.Update(payload.Id, payload)
}

func deleteChannelEvent(payload DeleteChannelPayload) {
	ChannelCache.Invalidate(payload.Id)
}

func updateMessageEvent(payload UpdateMessagePayload) {
	MessageCache.Update(payload.Id, payload.Message)
}

func deleteMessageEvent(payload DeleteMessagePayload) {
	MessageCache.Invalidate(payload.Id)
}

func updateRoleEvent(payload UpdateRolePayload) {
	RoleCache.Update(payload.Role.Id, payload.Role)
}

func deleteRoleEvent(payload DeleteRolePayload) {
	RoleCache.Invalidate(payload.RoleId)
}
