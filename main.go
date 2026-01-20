package main

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"ElainaBot/elaina"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const intents = discord.IntentGuildMessages | discord.IntentMessageContent

type Secrets struct {
	clientId     string
	clientSecret string
	token        string
}

func main() {
	secrets, err := loadSecrets()
	if err != nil {
		panic(err)
	}

	client, err := discord.CreateClient("ElainaBot", secrets.clientId, secrets.clientSecret, secrets.token, intents)
	if err != nil {
		panic(err)
	}

	if err = config.InitialiseConfig(); err != nil {
		panic(err)
	}

	elaina.RegisterEvents(&client.Events)
	elaina.RegisterCommands(client)

	deploy := flag.String("mode", "bot", "Set the running mode: deploy_commands=deploy application commands, deploy_db=deploy database")
	flag.Parse()

	switch *deploy {
	case "deploy_commands":
		client.DeployAllCommands()
	case "deploy_db":
		if err = elaina.InitialiseDatabase(); err != nil {
			panic(err)
		}
	default:
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

func loadSecrets() (secrets Secrets, err error) {
	if os.Getenv("ELAINA_DEBUG") == "true" {
		secrets.clientId = "1162820208315084921" // Devaina's client ID
		secrets.clientSecret = os.Getenv("DEVAINA_CLIENT_SECRET")
		secrets.token = os.Getenv("DEVAINA_TOKEN")
	} else {
		secrets.clientId = "1161747004712554656" // Elaina's client ID

		// Production secrets managed via docker compose secrets
		file, err := os.ReadFile("/run/secrets/elaina-secret")
		if err != nil {
			return secrets, err
		}
		secrets.clientSecret = string(file)

		file, err = os.ReadFile("/run/secrets/elaina-token")
		if err != nil {
			return secrets, err
		}
		secrets.token = string(file)
	}
	return
}
