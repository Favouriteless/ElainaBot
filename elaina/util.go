package elaina

import (
	"ElainaBot/discord"
)

func getMemberPerms(guild *discord.Guild, member *discord.GuildMember) (discord.Permissions, error) {
	// if guild.OwnerId == member.User.Id {
	// 	return 1<<64 - 1, nil
	// }
	//
	// everyone, err := discord.GetRole(guild.Id, guild.Id)
	// if err != nil {
	// 	return 0, err
	// }
	//
	// perms := everyone.Permissions
	//
	// for _, roleId := range member.Roles {
	// 	role, err := discord.GetRole(guild, roleId)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	perms = perms | role.Permissions
	// }
	// return perms, nil
	return 0, nil
}

func getMemberPermsInChannel(guild discord.Snowflake, member *discord.GuildMember, channel *discord.Channel) (discord.Permissions, error) {
	// perms, err := getMemberPerms(guild, member, client)
	// if err != nil {
	// 	return 0, err
	// }
	//
	// for _, over := range channel.PermissionOverwrites {
	// 	if over.Type == 1 && member.User.Id == over.Id { // User overwrite
	// 		perms = perms | over.Allow
	// 	}
	// }
	//
	// return perms | channel.Permission
	return 0, nil
}
