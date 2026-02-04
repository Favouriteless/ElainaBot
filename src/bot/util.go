package main

import (
	. "elaina-common"
	"elaina-common/restapi"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"time"
)

var customEmojiRegex = regexp.MustCompile("^<a?:.{2,}?:\\d{18,20}>$")

func getMemberPerms(guild Guild, member GuildMember, user Snowflake) (Permissions, error) {
	if guild.OwnerId == user {
		return 1<<64 - 1, nil
	}

	everyone, err := restapi.GetRole(guild.Id, guild.Id)
	if err != nil {
		return 0, err
	}
	if everyone == nil {
		return 0, errors.New("no @everyone role was found for guild: " + guild.Id.String())
	}

	perms := everyone.Permissions
	for _, roleId := range member.Roles {
		role, err := restapi.GetRole(guild.Id, roleId)
		if err != nil {
			return 0, err
		}
		perms |= role.Permissions
	}

	if perms&PermAdministrator == PermAdministrator {
		return 1<<64 - 1, nil
	}

	return perms, nil
}

func getMemberPermsInChannel(guild Guild, member GuildMember, user Snowflake, channel Channel) (Permissions, error) {
	perms, err := getMemberPerms(guild, member, user)
	if err != nil {
		return 0, err
	}

	if perms&PermAdministrator == PermAdministrator {
		return perms, nil // Admin permission discards overwrites
	}

	var oUser *Overwrite
	var allow Permissions // Role-specific allow/deny can be calculated in-place
	var deny Permissions

	for _, overwrite := range channel.PermissionOverwrites { // Sort the overwrites first so we don't iterate over them multiple times
		if overwrite.Id == user {
			oUser = &overwrite // User-specific overwrites get deferred to the end
		} else if overwrite.Id == guild.Id {
			perms &= ^overwrite.Deny // @everyone overwrite can be applied immediately
			perms |= overwrite.Allow
		} else if slices.Contains(member.Roles, overwrite.Id) {
			allow |= overwrite.Allow
			deny |= overwrite.Deny
		}
	}

	perms &= ^deny // Apply @role specific overwrites after @everyone
	perms |= allow
	if oUser != nil {
		perms &= ^oUser.Deny // Apply user overwrites last
		perms |= oUser.Allow
	}

	return perms, nil
}

func banUser(guild Snowflake, user User, reason string, deleteMessages int) error {
	banMsg := "You have been banned.\nReason: " + reason

	if dm, err := restapi.CreateDM(user.Id); err != nil { // Unlike timeout, the user MUST be notified before they leave the server, or the bot can't send a DM
		slog.Error("[Elaina] Failed to notify user of ban:", slog.String("user", user.Username), slog.String("error", err.Error()))
	} else if _, err = restapi.CreateMessage(dm.Id, banMsg, false); err != nil {
		slog.Error("[Elaina] Failed to notify user of ban:", slog.String("user", user.Username), slog.String("error", err.Error()))
	}

	if err := restapi.CreateBan(guild, user.Id, deleteMessages); err != nil {
		return errors.New("failed to create ban: " + err.Error())
	}

	slog.Info("[Elaina] Banned user:", slog.String("id", user.Id.String()), slog.String("reason", reason))
	return nil
}

func unbanUser(guild Snowflake, user Snowflake) error {
	if err := restapi.DeleteBan(guild, user); err != nil { // Unban can't create a DM because we can't assume the user still shares a guild with Elaina
		return errors.New("failed to unban user: " + err.Error())
	}
	slog.Info("[Elaina] Unbanned user: " + user.String())
	return nil
}

func timeoutUser(guild Snowflake, user User, duration time.Duration, reason string) error {
	expires := time.Now().Add(duration)

	go func() {
		timeoutMsg := fmt.Sprintf("You have been timed out until <t:%d>.\nReason: %s", expires.Unix(), reason)

		if dm, err := restapi.CreateDM(user.Id); err != nil {
			slog.Error("[Elaina] Failed to notify user of timeout:", slog.String("user", user.Username), slog.String("error", err.Error()))
		} else if _, err = restapi.CreateMessage(dm.Id, timeoutMsg, false); err != nil {
			slog.Error("[Elaina] Failed to notify user of timeout:", slog.String("user", user.Username), slog.String("error", err.Error()))
		}
	}()

	if err := restapi.ModifyGuildMember(guild, user.Id, ModifyGuildMemberPayload{CommunicationDisabledUntil: &Nullable[time.Time]{Value: expires}}); err != nil {
		return errors.New("failed to modify guild member: " + err.Error())
	}
	slog.Info("[Elaina] User timed out:", slog.String("id", user.Id.String()), slog.Float64("duration", duration.Seconds()), slog.String("reason", reason))
	return nil
}
