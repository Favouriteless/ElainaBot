package main

import (
	"ElainaBot/discord"
	"log"
	"os"
)

const intents = discord.IntentGuildMessages | discord.IntentMessageContent

func main() {
	client, err := discord.CreateClient("ElainaBot", intents)
	if err != nil {
		log.Fatal(err)
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
			Handler: func(data discord.ApplicationCommandData) bool {
				echo, err := data.OptionByName("string").AsString()
				if err != nil {
					log.Printf("Error completing echo: %v", err)
					return false
				}
				log.Println("Echo command: " + echo)
				return true
			},
		},
	}

	log.Println("Running as: " + mode)
	switch mode {
	case "--bot":
		registerEvents(&client.Events)
		if err = client.ConnectGateway(); err != nil {
			log.Fatal(err)
		}
		select { // Infinite select for now, this will handle CLI inputs in a real application
		}
	case "--deploy_commands":
		client.DeployAllCommands()
	}
}

func registerEvents(dsp *discord.EventDispatcher) {
	dsp.CreateMessage.Register(func(payload discord.CreateMessagePayload) {
		log.Println(payload.Content)
	})
}
