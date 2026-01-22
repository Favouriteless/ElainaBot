package elaina

import (
	"ElainaBot/config"
	"ElainaBot/discord"
	"log/slog"
	"regexp"
)

var elainaRegex = regexp.MustCompile("(?i)elaina")

func RegisterEvents() {
	discord.Events.CreateMessage.Register(logMessagesEvent, respondToNameEvent, banHoneypotEvent)
}

func logMessagesEvent(payload discord.CreateMessagePayload) {
	if payload.Author.Bot {
		return
	}
	slog.Info("[Elaina] Message received:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
}

func respondToNameEvent(payload discord.CreateMessagePayload) {
	if payload.Author.Bot {
		return
	}

	if elainaRegex.MatchString(payload.Content) {
		if err := discord.CreateReaction(payload.ChannelId, payload.Id, config.GetString(config.HelloEmoji)); err != nil {
			slog.Error("[Elaina] Could not say hello to " + payload.Author.Username + ": " + err.Error())
		} else {
			slog.Info("[Elaina] Saying hello to " + payload.Author.Username)
		}
	}
}

func banHoneypotEvent(payload discord.CreateMessagePayload) {
	honeypot := config.GetSnowflake(config.HoneyPotChannel)

	if honeypot == nil || payload.ChannelId != *honeypot || payload.Author.Bot || payload.GuildId == 0 { // We don't want to ban bots or people in DMs
		return
	}

	if payload.Member.Permissions&discord.PermAdministrator > 0 {
		slog.Info("[Elaina] Administrator typed in honeypot channel:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
		return
	}
}
