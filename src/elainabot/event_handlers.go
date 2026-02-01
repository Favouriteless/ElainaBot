package main

import (
	. "elaina-common"
	"log/slog"
	"regexp"
)

var elainaRegex = regexp.MustCompile("(?i)common")

func registerEvents() {
	Events.CreateMessage.Register(logMessagesEvent, respondToNameEvent, banHoneypotEvent)
}

func logMessagesEvent(payload CreateMessagePayload) error {
	if payload.Author.Bot {
		return nil
	}
	slog.Info("[Elaina] Message received:", slog.String("author", payload.Author.Username), slog.String("content", payload.Content))
	return nil
}

func respondToNameEvent(payload CreateMessagePayload) error {
	if payload.Author.Bot {
		return nil
	}

	if elainaRegex.MatchString(payload.Content) {
		if err := payload.CreateReaction(getConfig(HelloEmoji)); err != nil {
			slog.Error("[Elaina] Could not say hello to " + payload.Author.Username + ": " + err.Error())
		} else {
			slog.Info("[Elaina] Saying hello to " + payload.Author.Username)
		}
	}
	return nil
}

func banHoneypotEvent(payload CreateMessagePayload) error {
	if payload.Author.Bot || payload.GuildId == 0 {
		return nil
	}

	settings, err := GetGuildSettings(payload.GuildId)
	if err != nil {
		return err
	}

	if settings.HoneypotChannel == nil || payload.ChannelId != *settings.HoneypotChannel {
		return nil
	}

	guild, err := GetGuild(payload.GuildId)
	if err != nil {
		return err
	}

	channel, err := GetChannel(payload.ChannelId)
	if err != nil {
		return err
	}

	perms, err := getMemberPermsInChannel(*guild, *payload.Member, payload.Author.Id, *channel)
	if err != nil || perms&PermAdministrator > 0 || perms&PermModerateMembers > 0 {
		return err
	}

	if err = banUser(payload.GuildId, payload.Author, "You typed in the honeypot channel", 900); err != nil {
		return err
	}
	return nil
}
