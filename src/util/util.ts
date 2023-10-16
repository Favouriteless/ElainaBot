import { Guild } from "discord.js";

export async function getMember(guild: Guild, id: string) {
    let member = guild.members.cache.get(id);
    if(member === undefined)
        member = await guild.members.fetch(id);
    return member;
}