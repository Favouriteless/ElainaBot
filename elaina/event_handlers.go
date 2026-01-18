package elaina

import (
	"ElainaBot/discord"
	"log/slog"
	"regexp"
)

var elainaRegex = regexp.MustCompile("(?i)elaina")

func RegisterEvents(events *discord.EventDispatcher) {
	events.CreateMessage.Register(logMessagesEvent, respondToNameEvent)
}

func logMessagesEvent(payload discord.CreateMessagePayload, client *discord.Client) {
	slog.Info("Message received:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
}

func respondToNameEvent(payload discord.CreateMessagePayload, client *discord.Client) {
	if elainaRegex.MatchString(payload.Content) {
		err := client.CreateReaction(payload.ChannelId, payload.Id, "elainastare:1462288926663512274")
		if err != nil {
			slog.Error("Could not say hello to " + payload.Author.Username + ": " + err.Error())
		} else {
			slog.Info("Saying hello to " + payload.Author.Username)
		}
	}
}
