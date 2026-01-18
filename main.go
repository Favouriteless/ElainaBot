package main

import (
	"ElainaBot/discord"
	"ElainaBot/elaina"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const intents = discord.IntentGuildMessages | discord.IntentMessageContent

func main() {
	client, err := discord.CreateClient("ElainaBot", os.Getenv("ELAINA_CLIENT_ID"), os.Getenv("ELAINA_CLIENT_SECRET"), os.Getenv("ELAINA_TOKEN"), intents)
	if err != nil {
		panic(err)
	}

	elaina.RegisterEvents(&client.Events)
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

	deploy := flag.Bool("deploy_commands", false, "Deploy the bot's application commands")
	flag.Parse()

	if *deploy {
		client.DeployAllCommands()
	} else {
		if err = client.ConnectGateway(); err != nil {
			panic(err)
		}

		// Wait for a SIGINT or SIGTERM signal to gracefully shut down
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down...")
		client.CloseGateway(false)
	}
}
