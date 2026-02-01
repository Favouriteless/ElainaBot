package main

import (
	. "elaina-common"
	"fmt"
	"time"
)

var honeypotCommand = ApplicationCommand{
	Name:        "honeypot",
	Description: "Update the honeypot channel Elaina bans people in",
	Type:        CmdTypeChatInput,
	Permissions: PermAdministrator,
	Contexts:    []CommandContext{CmdContextGuild},
	Handler:     honeypotHandler,
	Options: []CommandOption{
		{
			Name:        "channel",
			Description: "Channel to ban users for typing in",
			Type:        CmdOptChannel,
			Required:    true,
		},
	},
}

var banCommand = ApplicationCommand{
	Name:        "ban",
	Description: "Ban a user",
	Type:        CmdTypeChatInput,
	Permissions: PermBan,
	Contexts:    []CommandContext{CmdContextGuild},
	Handler:     banHandler,
	Options: []CommandOption{
		{
			Name:        "user",
			Description: "User to ban",
			Type:        CmdOptUser,
			Required:    true,
		},
		{
			Name:        "reason",
			Description: "Reason for ban",
			Type:        CmdOptString,
			Required:    false,
		},
		{
			Name:        "delete_messages",
			Description: "All of the user's messages within the last X seconds will be deleted.",
			Type:        CmdOptInt,
			Required:    false,
			MinValue:    0,
			MaxValue:    604800,
		},
	},
}

var unbanCommand = ApplicationCommand{
	Name:        "unban",
	Description: "Unban a user",
	Type:        CmdTypeChatInput,
	Permissions: PermBan,
	Contexts:    []CommandContext{CmdContextGuild},
	Handler:     unbanHandler,
	Options: []CommandOption{
		{
			Name:        "user",
			Description: "User to unban",
			Type:        CmdOptUser,
			Required:    true,
		},
	},
}

var timeoutCommand = ApplicationCommand{
	Name:        "timeout",
	Description: "Timeout a user, preventing them from sending text messages or joining voice channels",
	Type:        CmdTypeChatInput,
	Permissions: PermModerateMembers,
	Contexts:    []CommandContext{CmdContextGuild},
	Handler:     timeoutHandler,
	Options: []CommandOption{
		{
			Name:        "user",
			Description: "User to timeout",
			Type:        CmdOptUser,
			Required:    true,
		},
		{
			Name:        "duration",
			Description: "Duration of timeout in seconds",
			Type:        CmdOptInt, // Should be int64
			Required:    true,
			MinValue:    1,
			MaxValue:    2419200,
		},
		{
			Name:        "reason",
			Description: "Reason of timeout",
			Type:        CmdOptString,
			Required:    false,
			MinLength:   1,
		},
	},
}

func honeypotHandler(params CommandParams) error {
	channel := params.GetOption("channel").AsSnowflake()

	settings, err := GetGuildSettings(params.GuildId)
	if err != nil {
		return err
	}
	settings.HoneypotChannel = &channel
	if err = CreateOrUpdateGuildSettings(params.GuildId, settings); err != nil {
		return err
	}

	return SendInteractionMessageResponse(Message{
		Content: fmt.Sprintf("Honey pot channel set to: <#%s>", channel.String()),
		Flags:   MsgFlagEphemeral,
	}, params.InteractionId, params.InteractionToken)
}

func banHandler(params CommandParams) error {
	userId := params.GetOption("user").AsSnowflake()
	reason := "No reason specified"
	del := 0

	if opt := params.GetOption("reason"); opt != nil {
		reason = opt.AsString()
	}
	if opt := params.GetOption("delete_messages"); opt != nil {
		del = opt.AsInt()
	}

	// Deferred response because there's a bunch of API calls during the ban flow
	if err := SendInteractionResponse(InteractionResponse{Type: RespTypeDeferredChannelMessage}, params.InteractionId, params.InteractionToken); err != nil {
		return err
	}

	user := params.Resolved.Users[userId]
	if err := banUser(params.GuildId, user, reason, del); err != nil {
		return err
	}

	return EditInteractionResponse(user.Username+" was banned.", params.InteractionToken)
}

func unbanHandler(params CommandParams) error {
	userId := params.GetOption("user").AsSnowflake()
	if err := unbanUser(params.GuildId, userId); err != nil {
		return err
	}

	return SendInteractionMessageResponse(Message{Content: userId.String() + " was unbanned."}, params.InteractionId, params.InteractionToken)
}

func timeoutHandler(params CommandParams) error {
	userId := params.GetOption("user").AsSnowflake()
	reason := "No reason specified"
	duration := time.Second * time.Duration(params.GetOption("duration").AsInt64())

	if opt := params.GetOption("reason"); opt != nil {
		reason = opt.AsString()
	}

	user := params.Resolved.Users[userId]
	if err := timeoutUser(params.GuildId, user, duration, reason); err != nil {
		return err
	}

	return SendInteractionMessageResponse(Message{Content: user.Username + " was timed out."}, params.InteractionId, params.InteractionToken)
}
