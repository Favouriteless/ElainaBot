package main

import (
	"ElainaBot/config"
	"ElainaBot/database"
	"ElainaBot/discord"
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
	dbUser       string
	dbPassword   string
	dbAddress    string
}

func main() {
	secrets := loadSecrets()
	if err := config.InitializeConfig(); err != nil {
		panic(err)
	}

	discord.Initialize("ElainaBot", secrets.clientId, secrets.clientSecret, secrets.token)

	RegisterEvents()
	RegisterCommands()

	deploy := flag.String("mode", "bot", "Update the running mode:\n- deploy_commands: Deploys application commands\n- deploy_db: Deploys/updates database schemas")
	commands := flag.String("commands", "", "Update the commands to deploy/delete when using the --mode=deploy_commands")
	flag.Parse()

	switch *deploy {
	case "deploy_commands":
		discord.DeployCommands(strings.Split(*commands, ",")...)
	case "deploy_db":

	case "bot":
		database.Deploy(secrets.dbUser, secrets.dbPassword, secrets.dbAddress) // Temporary measure to get the bot to auto update schemas

		conn := database.Connect(secrets.dbUser, secrets.dbPassword, secrets.dbAddress)
		defer conn.Close()

		handle := discord.ListenGateway(intents)

		go deleteExpiredBans()

		// Wait for a SIGINT or SIGTERM signal to gracefully shut down
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-sigChan:
				slog.Info("[Elaina] Shutting down...")
				handle.Close()
			case err := <-handle.Done:
				if err != nil {
					slog.Error("[Elaina] Gateway connection closed with an error: " + err.Error())
				} else {
					slog.Info("[Elaina] Gateway connection closed")
				}
				return
			}
		}
	default:
		slog.Error("[Elaina] Unknown execution mode: " + *deploy)
	}
}

func loadSecrets() Secrets {
	if os.Getenv("ELAINA_DEBUG") == "true" {
		return Secrets{
			clientId:     "1162820208315084921", // Devaina's client ID
			clientSecret: os.Getenv("DEVAINA_CLIENT_SECRET"),
			token:        os.Getenv("DEVAINA_TOKEN"),
			dbUser:       "devaina",
			dbPassword:   "devaina",
			dbAddress:    "localhost:3306",
		}
	}

	// Production secrets managed via docker compose secrets
	return Secrets{
		clientId:     "1161747004712554656", // Elaina's client ID
		clientSecret: dockerSecret("elaina-secret"),
		token:        dockerSecret("elaina-token"),
		dbUser:       dockerSecret("elaina-db-username"),
		dbPassword:   dockerSecret("elaina-db-password"),
		dbAddress:    dockerSecret("elaina-db-address"),
	}
}

func dockerSecret(fileName string) string {
	file, err := os.ReadFile("/run/secrets/" + fileName)
	if err != nil {
		panic(err) // We can't start the bot if secrets fail to load
	}
	return string(file)
}
