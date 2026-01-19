package elaina

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"errors"
	"log/slog"
	"regexp"

	"github.com/zmwangx/emojiregexp"
)

func RegisterCommands(client *discord.Client) {
	client.Commands = []*discord.ApplicationCommand{
		{
			Name:        "echo",
			Type:        discord.CmdTypeChatInput,
			Description: "Repeats what you said back to you",
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("string", "Echo... echo... echo...", true),
			},
			Handler: echoHandler,
		},
		{
			Name:        "sethelloemoji",
			Description: "Set the emoji Elaina uses to say hi",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("emoji", "Emoji to use as hello. If \"default\" is used, Elaina will reset the emoji", true),
			},
			Permissions: discord.PermAdministrator,
			Handler:     setHelloHandler,
		},
		{
			Name:        "setmacro",
			Description: "Set a macro for Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("key", "Key used to trigger the macro", true),
				discord.CmdOptString.Create("response", "The text Elaina will respond with", true),
			},
			Permissions: discord.PermAdministrator,
			Handler:     setMacroHandler,
		},
		{
			Name:        "deletemacro",
			Description: "Delete a macro from Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("key", "Key used to trigger the macro", true),
			},
			Permissions: discord.PermAdministrator,
			Handler:     deleteMacroHandler,
		},
		{
			Name:        "macro",
			Description: "Fetches a string by key from Elaina",
			Type:        discord.CmdTypeChatInput,
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("key", "Key used to trigger the macro", true),
			},
			Handler: useMacroHandler,
		},
	}
}

var customEmojiRegex = regexp.MustCompile("^<a?:.{2,}?:\\d{18,20}>$")

func setHelloHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string, client *discord.Client) error {
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

	_, err = client.SendInteractionResponse(discord.InteractionResponse{
		Type: discord.RespTypeChannelMessage,
		Data: discord.Message{Content: reply, Flags: discord.MsgFlagEphemeral},
	}, id, token)
	return err
}

func echoHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string, client *discord.Client) error {
	echo, err := data.OptionByName("string").AsString()
	if err != nil {
		return err
	}

	_, err = client.SendInteractionResponse(discord.InteractionResponse{
		Type: discord.RespTypeChannelMessage,
		Data: discord.Message{Content: echo, Flags: discord.MsgFlagEphemeral},
	}, id, token)
	return err
}

func setMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string, client *discord.Client) (err error) {
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

	_, err = client.SendInteractionResponse(discord.InteractionResponse{
		Type: discord.RespTypeChannelMessage,
		Data: discord.Message{Content: "Macro set!", Flags: discord.MsgFlagEphemeral},
	}, id, token)
	return err
}

func deleteMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string, client *discord.Client) error {
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

	_, err = client.SendInteractionResponse(discord.InteractionResponse{
		Type: discord.RespTypeChannelMessage,
		Data: discord.Message{Content: response, Flags: discord.MsgFlagEphemeral},
	}, id, token)
	return err
}

func useMacroHandler(data discord.ApplicationCommandData, id discord.Snowflake, token string, client *discord.Client) error {
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

	_, err = client.SendInteractionResponse(discord.InteractionResponse{
		Type: discord.RespTypeChannelMessage,
		Data: response,
	}, id, token)
	if err != nil {
		return err
	}
	return nil
}
