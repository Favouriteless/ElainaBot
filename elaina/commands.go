package elaina

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"errors"
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
