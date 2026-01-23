package main

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/zmwangx/emojiregexp"
)

func RegisterCommands() {
	discord.Commands = []*discord.ApplicationCommand{
		{
			Name:        "echo",
			Type:        discord.CmdTypeChatInput,
			Description: "Repeats what you said back to you",
			Options: []discord.CommandOption{
				{
					Name:        "string",
					Description: "Echo... echo... echo...",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
			Handler: echoHandler,
		},
		{
			Name:        "sethelloemoji",
			Description: "Update the emoji Elaina uses to say hi",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "emoji",
					Description: "Emoji to use as hello. If \"default\" is used, Elaina will reset the emoji",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
			Permissions: discord.PermAdministrator,
			Handler:     setHelloHandler,
		},
		{
			Name:        "setmacro",
			Description: "Update a macro for Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "key",
					Description: "Key used to trigger the macro",
					Type:        discord.CmdOptString,
					Required:    true,
				},
				{
					Name:        "response",
					Description: "The text Elaina will respond with",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
			Permissions: discord.PermAdministrator,
			Handler:     setMacroHandler,
		},
		{
			Name:        "deletemacro",
			Description: "Invalidate a macro from Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "key",
					Description: "Key used to trigger the macro",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
			Permissions: discord.PermAdministrator,
			Handler:     deleteMacroHandler,
		},
		{
			Name:        "macro",
			Description: "Fetches a string by key from Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "key",
					Description: "Key used to trigger the macro",
					Type:        discord.CmdOptString,
					Required:    true,
				},
			},
			Handler: useMacroHandler,
		},
		{
			Name:        "sethoneypot",
			Description: "Update the honeypot channel Elaina bans people in",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "channel",
					Description: "Channel to ban users for typing in",
					Type:        discord.CmdOptChannel,
					Required:    true,
				},
			},
			Permissions: discord.PermAdministrator,
			Handler:     setHoneypotHandler,
		},
		{
			Name:        "ban",
			Description: "Ban a user",
			Type:        discord.CmdTypeChatInput,
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
			Permissions: discord.PermBan,
			Handler:     banHandler,
		},
		{
			Name:        "unban",
			Description: "Ban a user",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				{
					Name:        "user",
					Description: "User to unban",
					Type:        discord.CmdOptUser,
					Required:    true,
				},
			},
			Permissions: discord.PermBan,
			Handler:     unbanHandler,
		},
	}
}

var customEmojiRegex = regexp.MustCompile("^<a?:.{2,}?:\\d{18,20}>$")

func setHelloHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	emoji := data.OptionByName("emoji").AsString()

	var reply string
	original := emoji
	updated := false

	if emoji == "default" {
		emoji = config.GetString(config.DefaultHelloEmoji)
		updated = true
	}
	if customEmojiRegex.MatchString(emoji) {
		emoji = emoji[1 : len(emoji)-1] // Strip the braces
		if len(emoji) > 2 && emoji[0:2] == "a:" {
			emoji = emoji[2:]
		}

		updated = true
	} else if emojiregexp.EmojiRegexp.MatchString(emoji) {
		updated = true
	}

	if updated {
		config.SetString(config.HelloEmoji, emoji)
		if err := config.SaveConfig(); err != nil {
			return err
		}
		reply = "Config has been updated! Hello emoji is now set to " + original
	} else {
		return errors.New(original + " is not a valid emoji")
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: reply, Flags: discord.MsgFlagEphemeral}, id, token)
}

func echoHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	echo := data.OptionByName("string").AsString()
	return discord.SendInteractionMessageResponse(discord.Message{Content: echo, Flags: discord.MsgFlagEphemeral}, id, token)
}

func setMacroHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) (err error) {
	macro := Macro{
		Key:      data.OptionByName("key").AsString(),
		Response: data.OptionByName("response").AsString(),
	}

	if err = SetMacro(macro); err != nil {
		return err
	}
	slog.Info("Macro set:", slog.String("key", macro.Key), slog.String("response", macro.Response))

	return discord.SendInteractionMessageResponse(discord.Message{Content: "Macro set!", Flags: discord.MsgFlagEphemeral}, id, token)
}

func deleteMacroHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	key := data.OptionByName("key").AsString()

	var response string
	if deleted, err := DeleteMacro(key); err != nil {
		return err
	} else if deleted {
		response = "Macro deleted"
		slog.Info("[Elaina] Macro deleted: \"" + key + "\"")
	} else {
		response = "No macro found for \"" + key + "\""
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: response, Flags: discord.MsgFlagEphemeral}, id, token)
}

func useMacroHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	key := data.OptionByName("key").AsString()
	macro, err := GetMacro(key)
	if err != nil {
		return err
	}

	var response discord.Message
	if macro != nil {
		response = discord.Message{Content: macro.Response}
	} else {
		response = discord.Message{Content: "No macro found for \"" + key + "\"", Flags: discord.MsgFlagEphemeral}
	}

	return discord.SendInteractionMessageResponse(response, id, token)
}

func setHoneypotHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	channel := data.OptionByName("channel").AsSnowflake()

	config.SetSnowflake(config.HoneyPotChannel, channel)
	if err := config.SaveConfig(); err != nil {
		return err
	}

	return discord.SendInteractionMessageResponse(discord.Message{
		Content: fmt.Sprintf("Honey pot channel set to: <#%s>", channel.String()),
		Flags:   discord.MsgFlagEphemeral,
	}, id, token)
}

func banHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	userId := data.OptionByName("user").AsSnowflake()
	duration := 0
	reason := "No reason specified"
	del := false

	if opt := data.OptionByName("duration"); opt != nil {
		duration = opt.AsInt()
	}
	if opt := data.OptionByName("reason"); opt != nil {
		reason = opt.AsString()
	}
	if opt := data.OptionByName("delete_messages"); opt != nil {
		del = opt.AsBool()
	}

	// Deferred response because there's a bunch of API calls during the ban flow
	err := discord.SendInteractionResponse(discord.InteractionResponse{Type: discord.RespTypeDeferredChannelMessage}, id, token)
	if err != nil {
		return err
	}

	user := data.ResolvedData.Users[userId]
	if err := banUser(guildId, user, duration, reason, del); err != nil {
		return err
	}

	return discord.EditInteractionResponse(user.Username+" was banned.", token)
}

func unbanHandler(data discord.ApplicationCommandData, guildId discord.Snowflake, id discord.Snowflake, token string) error {
	userId := data.OptionByName("user").AsSnowflake()
	if err := unbanUser(guildId, userId); err != nil {
		return err
	}
	return discord.SendInteractionMessageResponse(discord.Message{Content: userId.String() + " was unbanned."}, id, token)
}
