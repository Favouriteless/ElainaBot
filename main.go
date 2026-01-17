package main

import (
	"ElainaBot/discord"
	"log/slog"
	"os"
)

const intents = discord.IntentGuildMessages | discord.IntentMessageContent

func main() {
	client, err := discord.CreateClient("ElainaBot", intents)
	if err != nil {
		panic(err)
	}

	var mode string
	if len(os.Args) > 1 {
		mode = os.Args[1]
	} else {
		mode = "--bot"
	}

	client.Commands = []*discord.ApplicationCommand{
		{
			Name:        "echo",
			Type:        discord.CmdTypeChatInput,
			Description: "String testing command",
			Options: []discord.CommandOption{
				discord.CmdOptString.Create("string", "Testing string option for Devaina", true),
			},
			Handler: func(data discord.ApplicationCommandData, id discord.Snowflake, token string) error {
				echo, err := data.OptionByName("string").AsString()
				if err != nil {
					return err
				}

				resp := discord.InteractionResponse{Type: discord.RespTypeChannelMessage, Data: discord.Message{Content: echo, Flags: discord.MsgFlagEphemeral}}
				_, err = client.SendInteractionResponse(id, token, resp)
				return err
			},
		},
	}

	slog.Info("Running as: " + mode)
	switch mode {
	case "--bot":

		client.Events.CreateMessage.Register(func(payload discord.CreateMessagePayload) {
			slog.Info("Message received:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
		})

		if err = client.ConnectGateway(); err != nil {
			panic(err)
		}
		select { // Infinite select for now, this will handle CLI inputs in a real application
		}
	case "--deploy_commands":
		client.DeployAllCommands()
	}
}
