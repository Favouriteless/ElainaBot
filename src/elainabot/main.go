package main

import (
	. "elaina-common"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const intents = IntentGuildMessages | IntentMessageContent

var botSecrets struct {
	dbUser     string
	dbPassword string
	dbAddress  string
}

func main() {
	loadSecrets()

	if err := initializeConfig(); err != nil {
		panic(err)
	}

	registerEvents()
	registerCommands()

	deploy := flag.String("mode", "", "Update the running mode:\n- deploy_commands: Deploys application toDeploy\n- deploy_db: Deploys/updates database schemas")
	commands := flag.String("commands", "", "A comma separated list of commands to deploy")
	flag.Parse()

	switch *deploy {
	case "deploy_commands":
		if *commands == "" {
			DeployCommands(Commands)
			return
		}
		names := strings.Split(*commands, ",")

		toDeploy := make(CommandCollection, 0, len(names))
		for _, name := range names {
			cmd := Commands.GetCommand(name)
			if cmd == nil {
				slog.Error("[Elaina] Tried to deploy nonexistent command: " + name)
				return
			}
			toDeploy = append(toDeploy, cmd)
		}
		DeployCommands(toDeploy) // TODO: Both the deploy command and deploy db functions are fundamentally incompatible with sharding. These should be built into a separate util application
	case "deploy_db":
		DeployDatabase(botSecrets.dbUser, botSecrets.dbPassword, botSecrets.dbAddress)
	case "bot":
		DeployDatabase(botSecrets.dbUser, botSecrets.dbPassword, botSecrets.dbAddress) // Temporary measure to get the bot to auto update schemas

		db := ConnectDatabase(botSecrets.dbUser, botSecrets.dbPassword, botSecrets.dbAddress)
		defer db.Close()

		handle := listenGateway(intents)

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

func registerCommands() {
	Commands = []*ApplicationCommand{
		&echoCommand, &macroCommand, &editMacroCommand, &honeypotCommand, &banCommand, &unbanCommand, &timeoutCommand,
	}
}

func loadSecrets() {
	if os.Getenv("ELAINA_DEBUG") == "true" {
		CommonSecrets.Id = "1162820208315084921" // Devaina's client ID
		CommonSecrets.Secret = os.Getenv("DEVAINA_CLIENT_SECRET")
		CommonSecrets.BotToken = os.Getenv("DEVAINA_TOKEN")

		botSecrets.dbUser = "devaina"
		botSecrets.dbPassword = "devaina"
		botSecrets.dbAddress = "localhost:3306"
	} else {
		// Production secrets managed via docker compose secrets
		CommonSecrets.Id = "1161747004712554656" // Elaina's client ID
		CommonSecrets.Secret = dockerSecret("elaina-secret")
		CommonSecrets.Secret = os.Getenv("DEVAINA_CLIENT_SECRET")

		botSecrets.dbUser = dockerSecret("elaina-db-username")
		botSecrets.dbPassword = dockerSecret("elaina-db-password")
		botSecrets.dbAddress = dockerSecret("elaina-db-address")
	}
}

func dockerSecret(fileName string) string {
	file, err := os.ReadFile("/run/secrets/" + fileName)
	if err != nil {
		panic(err) // We can't start the bot if secrets fail to load
	}
	return string(file)
}
