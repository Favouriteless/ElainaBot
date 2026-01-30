package main

import (
	"ElainaBot/database"
	"ElainaBot/discord"
	"ElainaBot/elaina"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
)

func RegisterCommands() {
	discord.Commands = []*discord.ApplicationCommand{
		{
			Name:        "echo",
			Type:        discord.CmdTypeChatInput,
			Description: "Repeats what you said back to you",
			Handler:     echoHandler,
			Options: []discord.CommandOption{
				{
					Name:        "string",
					Description: "Echo... echo... echo...",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
		},
		{
			Name:        "editmacro",
			Description: "Set or delete a macro",
			Type:        discord.CmdTypeChatInput,
			Permissions: discord.PermAdministrator,
			Contexts:    []discord.CommandContext{discord.CmdContextGuild},
			Options: []discord.CommandOption{
				{
					Name:        "set",
					Description: "Set a macro",
					Type:        discord.CmdOptSubcommand,
					Handler:     macroSetHandler,
					Options: []discord.CommandOption{
						{
							Name:        "keyword",
							Description: "Keyword used to trigger the macro",
							Type:        discord.CmdOptString,
							Required:    true,
						},
						{
							Name:        "response",
							Description: "The text Elaina will respond with",
							Type:        discord.CmdOptString,
							MinLength:   1,
							MaxLength:   280,
							Required:    true,
						},
					},
				},
				{
					Name:        "delete",
					Description: "Delete a macro",
					Type:        discord.CmdOptSubcommand,
					Handler:     macroDeleteHandler,
					Options: []discord.CommandOption{
						{
							Name:        "keyword",
							Description: "Keyword used to trigger the macro",
							Type:        discord.CmdOptString,
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "macro",
			Description: "Macros are a handy way to save a message and fetch it using a keyword",
			Type:        discord.CmdTypeChatInput,
			Contexts:    []discord.CommandContext{discord.CmdContextGuild},
			Handler:     macroUseHandler,
			Options: []discord.CommandOption{
				{
					Name:        "keyword",
					Description: "Keyword used to trigger the macro",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
		},
		{
			Name:        "honeypot",
			Description: "Update the honeypot channel Elaina bans people in",
			Type:        discord.CmdTypeChatInput,
			Permissions: discord.PermAdministrator,
			Contexts:    []discord.CommandContext{discord.CmdContextGuild},
			Handler:     honeypotHandler,
			Options: []discord.CommandOption{
				{
					Name:        "channel",
					Description: "Channel to ban users for typing in",
					Type:        discord.CmdOptChannel,
					Required:    true,
				},
			},
		},
		{
			Name:        "ban",
			Description: "Ban a user",
			Type:        discord.CmdTypeChatInput,
			Permissions: discord.PermBan,
			Contexts:    []discord.CommandContext{discord.CmdContextGuild},
			Handler:     banHandler,
			Options: []discord.CommandOption{
				{
					Name:        "user",
					Description: "User to ban",
					Type:        discord.CmdOptUser,
					Required:    true,
				},
				{
					Name:        "duration",
					Description: "Duration of ban in seconds",
					Type:        discord.CmdOptInt,
					Required:    false,
				},
				{
					Name:        "reason",
					Description: "Reason for ban",
					Type:        discord.CmdOptString,
					Required:    false,
				},
				{
					Name:        "delete_messages",
					Description: "If true, the last day of the users messages will be deleted",
					Type:        discord.CmdOptBool,
					Required:    false,
				},
			},
		},
		{
			Name:        "unban",
			Description: "Ban a user",
			Type:        discord.CmdTypeChatInput,
			Permissions: discord.PermBan,
			Contexts:    []discord.CommandContext{discord.CmdContextGuild},
			Handler:     unbanHandler,
			Options: []discord.CommandOption{
				{
					Name:        "user",
					Description: "User to unban",
					Type:        discord.CmdOptUser,
					Required:    true,
				},
			},
		},
	}
}

var customEmojiRegex = regexp.MustCompile("^<a?:.{2,}?:\\d{18,20}>$")

func echoHandler(params discord.CommandParams) error {
	echo := params.GetOption("string").AsString()
	return discord.SendInteractionMessageResponse(discord.Message{Content: echo, Flags: discord.MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroSetHandler(params discord.CommandParams) (err error) {
	macro := elaina.Macro{
		Guild:    params.GuildId,
		Key:      params.GetOption("keyword").AsString(),
		Response: params.GetOption("response").AsString(),
	}

	if err = database.CreateOrUpdateMacro(macro); err != nil {
		return err
	}
	slog.Info("Macro set:", slog.String("key", macro.Key), slog.String("response", macro.Response))

	return discord.SendInteractionMessageResponse(discord.Message{Content: "Macro set!", Flags: discord.MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroDeleteHandler(params discord.CommandParams) error {
	key := params.GetOption("keyword").AsString()

	var response string
	if deleted, err := database.DeleteMacro(params.GuildId, key); err != nil {
		return err
	} else if deleted {
		response = "Macro deleted"
		slog.Info("[Elaina] Macro deleted: \"" + key + "\"")
	} else {
		response = "No macro found for \"" + key + "\""
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: response, Flags: discord.MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroUseHandler(params discord.CommandParams) error {
	key := params.GetOption("keyword").AsString()
	macro, err := database.GetMacro(params.GuildId, key)
	if err != nil {
		return err
	}

	var response discord.Message
	if macro != nil {
		response = discord.Message{Content: macro.Response}
	} else {
		response = discord.Message{Content: "No macro found for \"" + key + "\"", Flags: discord.MsgFlagEphemeral}
	}

	return discord.SendInteractionMessageResponse(response, params.InteractionId, params.InteractionToken)
}

func honeypotHandler(params discord.CommandParams) error {
	channel := params.GetOption("channel").AsSnowflake()

	settings, err := database.GetGuildSettings(params.GuildId)
	if err != nil {
		return err
	}
	settings.HoneypotChannel = &channel
	if err = database.CreateOrUpdateGuildSettings(params.GuildId, settings); err != nil {
		return err
	}

	return discord.SendInteractionMessageResponse(discord.Message{
		Content: fmt.Sprintf("Honey pot channel set to: <#%s>", channel.String()),
		Flags:   discord.MsgFlagEphemeral,
	}, params.InteractionId, params.InteractionToken)
}

func banHandler(params discord.CommandParams) error {
	userId := params.GetOption("user").AsSnowflake()
	duration := 0
	reason := "No reason specified"
	del := false

	if opt := params.GetOption("duration"); opt != nil {
		duration = opt.AsInt()
	}
	if opt := params.GetOption("reason"); opt != nil {
		reason = opt.AsString()
	}
	if opt := params.GetOption("delete_messages"); opt != nil {
		del = opt.AsBool()
	}

	// Deferred response because there's a bunch of API calls during the ban flow
	if err := discord.SendInteractionResponse(discord.InteractionResponse{Type: discord.RespTypeDeferredChannelMessage}, params.InteractionId, params.InteractionToken); err != nil {
		return err
	}
	user := params.Resolved.Users[userId]
	if err := banUser(params.GuildId, user, duration, reason, del); err != nil {
		return err
	}

	return discord.EditInteractionResponse(user.Username+" was banned.", params.InteractionToken)
}

func unbanHandler(params discord.CommandParams) error {
	userId := params.GetOption("user").AsSnowflake()
	if err := unbanUser(params.GuildId, userId); err != nil {
		return err
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: userId.String() + " was unbanned."}, params.InteractionId, params.InteractionToken)
}

func testGroupHandler(params discord.CommandParams) error {
	return errors.New("group not implemented")
}

func testSubHandler(params discord.CommandParams) error {
	return errors.New("sub not implemented")
}
