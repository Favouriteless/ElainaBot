package main

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"ElainaBot/elaina"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

	discord.Initialize("ElainaBot", secrets.clientId, secrets.clientSecret, secrets.token)

	if err = config.InitializeConfig(); err != nil {
		panic(err)
	}

	elaina.RegisterEvents()
	elaina.RegisterCommands()

	deploy := flag.String("mode", "bot", "Update the running mode: deploy_commands=deploy application commands, deploy_db=deploy database")
	commands := flag.String("commands", "", "Update the commands to deploy when using the --mode=deploy_commands")
	flag.Parse()

	switch *deploy {
	case "deploy_commands":
		discord.DeployCommands(strings.Split(*commands, ",")...)
	case "deploy_db":
		if err = elaina.InitializeDatabase(); err != nil {
			panic(err)
		}
	default:
		closeGateway := make(chan interface{}, 1)
		discord.ListenGateway(intents, closeGateway)

		// Wait for a SIGINT or SIGTERM signal to gracefully shut down
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		closeGateway <- true
		slog.Info("Shutting down...")
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
		file, err := os.ReadFile("secrets/elaina-secret")
		if err != nil {
			return secrets, err
		}
		secrets.clientSecret = string(file)

		file, err = os.ReadFile("secrets/elaina-token")
		if err != nil {
			return secrets, err
		}
		secrets.token = string(file)
	}
	return
}
