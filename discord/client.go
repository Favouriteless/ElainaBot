package discord

import (
	"encoding/json"
	"log"
	"os"
)

const apiUrl = "https://discord.com/api/v10/"

// Secrets represents the auth details of the discord client, all methods interfacing with Discord API will require a
// secrets to authenticate with.
type Secrets struct {
	id     string
	secret string
}

func LoadSecrets() *Secrets {
	log.Println("Loading secrets...")
	contents, err := os.ReadFile("data/secrets.json")
	if err != nil {
		log.Fatal("Error reading secrets")
	}

	var secrets struct { // This is a bit cursed, but we REALLY don't want these fields to be exported
		Id     string
		Secret string
	}
	err = json.Unmarshal(contents, &secrets)
	if err != nil {
		log.Fatal("Error parsing secrets")
	}
	log.Println("Finished loading secrets!")
	return &Secrets{id: secrets.Id, secret: secrets.Secret}
}
