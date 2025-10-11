package main

import (
	"ElainaBot/discord"
	"log"
)

const intents = discord.IGuildMessages | discord.IMessageContent

func main() {
	client, err := discord.CreateClient("ElainaBot", intents)
	if err != nil {
		log.Fatal(err)
	}
	registerEvents(&client.Events)

	if err = client.ConnectGateway(); err != nil {
		log.Fatal(err)
	}
	select {}
}

func registerEvents(dispatch *discord.EventDispatcher) {
	dispatch.CreateMessage.Register(func(payload discord.CreateMessagePayload) {
		log.Println(payload.Content)
	})
}
