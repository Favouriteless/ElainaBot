package main

import (
	"ElainaBot/discord"
	"io"
	"log"
	"os"
)

const intents = discord.IntentGuildMessages | discord.IntentMessageContent

func main() {
	client, err := discord.CreateClient("ElainaBot", intents)
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]

	var mode string
	if len(args) > 0 {
		mode = args[0]
	} else {
		mode = "--bot"
	}

	log.Println("Running as " + mode)
	switch mode {
	case "--bot":
		registerEvents(&client.Events)
		if err = client.ConnectGateway(); err != nil {
			log.Fatal(err)
		}
		select {}
	case "--deploy_commands":
		deployAppCommands(client)
	}
}

func registerEvents(dispatch *discord.EventDispatcher) {
	dispatch.CreateMessage.Register(func(payload discord.CreateMessagePayload) {
		log.Println(payload.Content)
	})
}

func deployAppCommands(client *discord.Client) {
	test := discord.CreateGlobalSlashCommand("test2", "Testing command added for Devaina", 0)
	test.Permissions(discord.PermAdministrator)
	test.Option(discord.OptBool.Create("test_param", "Testing param for command options", true))

	registerAppCommands(client, test)
}

func registerAppCommands(client *discord.Client, builders ...*discord.ApplicationCommandBuilder) {
	for _, builder := range builders {
		func() {
			com := builder.Build()
			resp, err := client.CreateOrUpdateApplicationCommand(com, 3)
			if err != nil {
				log.Printf("Error registering command \"%s\": %s", com.Name, err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case 200:
				log.Printf("Command \"%s\" already exists, it was updated", com.Name)
			case 201:
				log.Printf("Command \"%s\" added successfully", com.Name)
			default:
				log.Printf("Command \"%s\" could not be created: %s", com.Name, resp.Status)
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err) // Should never be hit
				}
				log.Println(string(body))
			}
		}()
	}
}
