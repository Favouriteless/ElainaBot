import { Events, Guild, Message, MessageReaction, PartialMessage, PartialMessageReaction, PartialUser, User } from "discord.js";
import { Client } from "../elaina";
import { deleteReactionRoles, getReactionRole } from "./db/db";

export function registerReactionRoleListeners(client: Client) {

    client.on(Events.MessageReactionAdd, async (reaction: MessageReaction | PartialMessageReaction, user: User | PartialUser) => {
        try {
            if(reaction.partial)
                await reaction.fetch();
            if(user.partial)
                await user.fetch();
        }
        catch(error) {
            console.error(`Could not retrieve reaction on message: ${error}`);
            return;
        }

        getReactionRole(reaction.message.id, reaction.emoji.toString()).then(reactionRole => {
            if(reactionRole !== undefined) { // If reaction role is appropriate for this message and emoji
                reaction.message.guild?.roles.fetch(reactionRole.roleId).then(async (role) => {
                    if(role !== undefined && role != null) {
                        if(reaction.message.guild)
                            getMember(reaction.message.guild, user.id).then(member => member.roles.add(reactionRole.roleId));
                    }
                });
            }
        });
    });

    client.on(Events.MessageReactionRemove, async (reaction: MessageReaction | PartialMessageReaction, user: User | PartialUser) => {
        try {
            if(reaction.partial)
                await reaction.fetch();
            if(user.partial)
                await user.fetch();
        }
        catch(error) {
            console.error(`Could not retrieve reaction on message: ${error}`);
            return;
        }

        getReactionRole(reaction.message.id, reaction.emoji.toString()).then(reactionRole => {
            if(reactionRole !== undefined) { // If reaction role is appropriate for this message and emoji
                reaction.message.guild?.roles.fetch(reactionRole.roleId).then(async (role) => {
                    if(role !== undefined && role != null) {
                        if(reaction.message.guild)
                            getMember(reaction.message.guild, user.id).then(member => member.roles.remove(reactionRole.roleId));
                    }
                });
            }
        });
    });

    client.on(Events.MessageDelete, async (message: Message | PartialMessage) => {
        const result = await deleteReactionRoles(message.id); // Clean up reaction roles for messages which get deleted.
        if(result.length > 0)
            console.log(`Deleted ${result.length} reaction role entries.`);
    });
    
}

async function getMember(guild: Guild, id: string) {
    let member = guild.members.cache.get(id);
    if(member === undefined)
        member = await guild.members.fetch(id);
    return member;
}