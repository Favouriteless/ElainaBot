package main

import (
	"ElainaBot/database"
	"ElainaBot/discord"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

func getMemberPerms(guild discord.Guild, member discord.GuildMember, user discord.Snowflake) (discord.Permissions, error) {
	if guild.OwnerId == user {
		return 1<<64 - 1, nil
	}

	everyone, err := guild.GetRole(guild.Id)
	if err != nil {
		return 0, err
	}
	if everyone == nil {
		return 0, errors.New("no @everyone role was found for guild: " + guild.Id.String())
	}

	perms := everyone.Permissions
	for _, roleId := range member.Roles {
		role, err := guild.GetRole(roleId)
		if err != nil {
			return 0, err
		}
		perms |= role.Permissions
	}

	if perms&discord.PermAdministrator == discord.PermAdministrator {
		return 1<<64 - 1, nil
	}

	return perms, nil
}

func getMemberPermsInChannel(guild discord.Guild, member discord.GuildMember, user discord.Snowflake, channel discord.Channel) (discord.Permissions, error) {
	perms, err := getMemberPerms(guild, member, user)
	if err != nil {
		return 0, err
	}

	if perms&discord.PermAdministrator == discord.PermAdministrator {
		return perms, nil // Admin permission discards overwrites
	}

	var oUser *discord.Overwrite
	var allow discord.Permissions // Role-specific allow/deny can be calculated in-place
	var deny discord.Permissions

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

func banUser(guild discord.Snowflake, user discord.User, duration int, reason string, deleteMessages bool) (err error) {
	expires := time.Now().Add(time.Second * time.Duration(duration)).Unix()

	var banMsg string
	if duration == 0 {
		banMsg = "You have been permanently banned.\nReason: " + reason
	} else {
		banMsg = fmt.Sprintf("You have been banned until <t:%d>.\nReason: %s", expires, reason)
	}

	if dm, err := user.CreateDM(); err != nil {
		slog.Error("[Elaina] Failed to create DM channel: " + err.Error())
	} else if _, err = dm.CreateMessage(banMsg, false); err != nil {
		slog.Error("[Elaina] Failed to create DM message: " + err.Error())
	}

	if err = database.CreateOrUpdateBan(guild, user.Id, expires, reason); err != nil {
		return err
	}
	if deleteMessages {
		err = discord.CreateBan(guild, user.Id, 86400)
	} else {
		err = discord.CreateBan(guild, user.Id, 0)
	}

	slog.Info("[Elaina] Banned user:", slog.String("id", user.Id.String()), slog.Int("duration", duration), slog.String("reason", reason))
	return err
}

func unbanUser(guild discord.Snowflake, user discord.Snowflake) error {
	if err := discord.DeleteBan(guild, user); err != nil {
		return err
	}
	err := database.DeleteBan(guild, user)
	if err == nil {
		slog.Info("[Elaina] Unbanned user: " + user.String())
	}
	return err
}

func deleteExpiredBans() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	// TODO: This global delete bans routine doesn't work with sharding
	for {
		now := <-ticker.C
		unix := now.Unix()

		bans, err := database.GetExpiredBans(unix)
		if err != nil {
			slog.Error("[Elaina] Failed to check bans: " + err.Error())
			return
		}

		for _, ban := range bans {
			if err = unbanUser(ban.Guild, ban.User); err != nil {
				slog.Error("[Elaina] Failed to unban user: " + err.Error())
			}
		}
	}
}
