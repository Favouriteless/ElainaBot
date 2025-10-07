package main

import (
	"ElainaBot/discord"
	"fmt"
	"log"
	"time"
)

func main() {
	client, err := discord.CreateClient("ElainaBot", time.Second*5)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartGatewaySession()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Passed initialisation")
}
