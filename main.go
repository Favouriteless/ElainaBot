package main

import (
	"ElainaBot/discord"
	"fmt"
	"log"
)

func main() {
	_, err := discord.CreateClient()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Passed initialisation")
}
