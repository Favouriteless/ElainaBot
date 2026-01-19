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

func main() {
	client, err := discord.CreateClient("ElainaBot", os.Getenv("ELAINA_CLIENT_ID"), os.Getenv("ELAINA_CLIENT_SECRET"), os.Getenv("ELAINA_TOKEN"), intents)
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
	case "test":
		macro, err := elaina.GetMacro("test")
		if err != nil {
			panic(err)
		}
		slog.Info(macro.Response)
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
