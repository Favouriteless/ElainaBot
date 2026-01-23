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

func logMessagesEvent(payload discord.CreateMessagePayload) error {
	if payload.Author.Bot {
		return nil
	}
	slog.Info("[Elaina] Message received:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
	return nil
}

func respondToNameEvent(payload discord.CreateMessagePayload) error {
	if payload.Author.Bot {
		return nil
	}

	if elainaRegex.MatchString(payload.Content) {
		if err := payload.CreateReaction(config.GetString(config.HelloEmoji)); err != nil {
			slog.Error("[Elaina] Could not say hello to " + payload.Author.Username + ": " + err.Error())
		} else {
			slog.Info("[Elaina] Saying hello to " + payload.Author.Username)
		}
	}
	return nil
}

func banHoneypotEvent(payload discord.CreateMessagePayload) error {
	honeypot := config.GetSnowflake(config.HoneyPotChannel)
	if honeypot == nil || payload.ChannelId != *honeypot || payload.Author.Bot || payload.GuildId == 0 {
		return nil // We don't want to ban bots or people in the wrong channel
	}
	guild, err := discord.GetGuild(payload.GuildId)
	if err != nil {
		return err
	}

	channel, err := discord.GetChannel(payload.ChannelId)
	if err != nil {
		return err
	}

	perms, err := getMemberPermsInChannel(*guild, *payload.Member, payload.Author.Id, *channel)
	if err != nil || perms&discord.PermAdministrator > 0 || perms&discord.PermModerateMembers > 0 {
		return err
	}

	if err = discord.CreateBan(payload.GuildId, payload.Author.Id, 604800); err != nil {
		return err
	}
	slog.Info("[Honeypot] banned user: " + payload.Author.Username)
	return nil
}
