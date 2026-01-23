package discord

import (
	"encoding/json"
	"log/slog"
)

// interactionCreateEvent dispatches ApplicationCommands
func interactionCreateEvent(payload InteractionCreatePayload) error { // Built-in event handler for dispatching application Commands
	if payload.Type != 2 { // https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-data
		return nil
	}

	var c ApplicationCommandData
	if err := json.Unmarshal(*payload.Data, &c); err != nil {
		slog.Error("[Command] Failed to parse application command data: " + err.Error())
		return nil
	}

	for _, command := range Commands {
		if c.Name == command.Name {
			slog.Info("[Command] Dispatching application command: " + c.Name)

			if err := command.Handler(c, payload.Id, payload.Token); err != nil {
				slog.Error("[Command] Error executing application command: ", slog.String("command", c.Name), slog.String("error", err.Error()))
				_ = SendInteractionResponse(InteractionResponse{
					Type: RespTypeChannelMessage,
					Data: Message{Content: "An error occurred executing this command: " + err.Error(), Flags: MsgFlagEphemeral},
				}, payload.Id, payload.Token)
			}
			return nil
		}
	}
	slog.Warn("[Command] Application command was dispatched but no handler was found: " + c.Name)
	return nil
}

func updateChannelEvent(payload UpdateChannelPayload) error {
	ChannelCache.Update(payload.Id, payload)
	return nil
}

func deleteChannelEvent(payload DeleteChannelPayload) error {
	ChannelCache.Invalidate(payload.Id)
	return nil
}

func updateMessageEvent(payload UpdateMessagePayload) error {
	MessageCache.Update(payload.Id, payload.Message)
	return nil
}

func deleteMessageEvent(payload DeleteMessagePayload) error {
	MessageCache.Invalidate(payload.Id)
	return nil
}

func updateRoleEvent(payload UpdateRolePayload) error {
	RoleCache.Update(payload.Role.Id, payload.Role)
	return nil
}

func deleteRoleEvent(payload DeleteRolePayload) error {
	RoleCache.Invalidate(payload.RoleId)
	return nil
}

func updateGuildEvent(payload UpdateGuildPayload) error {
	GuildCache.Update(payload.Id, payload)
	return nil
}

func deleteGuildEvent(payload DeleteGuildPayload) error {
	GuildCache.Invalidate(payload.Id)
	return nil
}
