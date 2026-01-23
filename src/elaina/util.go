package elaina

import (
	"ElainaBot/discord"
	"errors"
	"slices"
)

func getMemberPerms(guild discord.Guild, member discord.GuildMember, user discord.Snowflake) (discord.Permissions, error) {
	if guild.OwnerId == user {
		return 1<<63 - 1, nil
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
		return 1<<63 - 1, nil
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
