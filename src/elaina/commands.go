package elaina

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
	}
}

var customEmojiRegex = regexp.MustCompile("^<a?:.{2,}?:\\d{18,20}>$")

func setHelloHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
	emoji, err := data.OptionByName("emoji").AsString()
	if err != nil {
		return err
	}

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
		config.Set(config.HelloEmoji, emoji)
		if err = config.SaveConfig(); err != nil {
			return err
		}
		reply = "Config has been updated! Hello emoji is now set to " + original
	} else {
		return errors.New(original + " is not a valid emoji")
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: reply, Flags: discord.MsgFlagEphemeral}, id, token)
}

func echoHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
	echo, err := data.OptionByName("string").AsString()
	if err != nil {
		return err
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: echo, Flags: discord.MsgFlagEphemeral}, id, token)
}

func setMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) (err error) {
	var macro Macro
	if macro.Key, err = data.OptionByName("key").AsString(); err != nil {
		return err
	}
	if macro.Response, err = data.OptionByName("response").AsString(); err != nil {
		return err
	}
	if err = SetMacro(macro); err != nil {
		return err
	}
	slog.Info("Macro set:", slog.String("key", macro.Key), slog.String("response", macro.Response))

	return discord.SendInteractionMessageResponse(discord.Message{Content: "Macro set!", Flags: discord.MsgFlagEphemeral}, id, token)
}

func deleteMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
	key, err := data.OptionByName("key").AsString()
	if err != nil {
		return err
	}

	deleted, err := DeleteMacro(key)

	var response string
	if deleted > 0 {
		response = "Macro deleted"
		slog.Info("Macro deleted:", slog.String("key", key))
	} else {
		response = "No macro found for \"" + key + "\""
	}

	return discord.SendInteractionMessageResponse(discord.Message{Content: response, Flags: discord.MsgFlagEphemeral}, id, token)
}

func useMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
	key, err := data.OptionByName("key").AsString()
	if err != nil {
		return err
	}
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

func setHoneypotHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
	channel, err := data.OptionByName("channel").AsSnowflake()
	if err != nil {
		return err
	}

	fetched, err := discord.GetChannel(channel)
	if err != nil {
		return err
	}
	if fetched == nil {
		return discord.SendInteractionMessageResponse(discord.Message{
			Content: "Could not find channel: " + channel.String(),
			Flags:   discord.MsgFlagEphemeral,
		}, id, token)
	}

	config.Set(config.HoneyPotChannel, channel)
	if err = config.SaveConfig(); err != nil {
		return err
	}

	return discord.SendInteractionMessageResponse(discord.Message{
		Content: fmt.Sprintf("Honey pot channel set to: <#%s>", channel),
		Flags:   discord.MsgFlagEphemeral,
	}, id, token)
}
