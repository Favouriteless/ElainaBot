import { Events, BaseInteraction, Message, GuildMember, MessageReaction, User, PartialMessageReaction, PartialUser } from "discord.js";
import { db, updateReply } from "./db/db";
import { Client } from "../elaina";
import { config } from "./config";

export function registerListeners(client: Client) {

    client.once(Events.ClientReady, c => {
        console.log(`Ready! Logged in as ${c.user.tag}`);
    });

	client.on(Events.InteractionCreate, async (interaction: BaseInteraction) => {
        if (!interaction.isChatInputCommand()) return;
        
        const command = (interaction.client as Client).commands.get(interaction.commandName);

        if (!command) {
            console.error(`No command matching ${interaction.commandName} was found.`);
            return;
        }

        try {
            await command.execute(interaction);
        }
        catch (error) {
            console.error(error);
            if (interaction.replied || interaction.deferred)
                await interaction.followUp({ content: 'There was an error while executing this command!', ephemeral: true });
            else
                await interaction.reply({ content: 'There was an error while executing this command!', ephemeral: true });
        }
    });

    client.on(Events.MessageCreate, async (message: Message) => {
        if (message.author.bot)
            return;

        let words = message.content.match(/(\?*[\w\-]+)/g);
        if(words == null)
            return;
        
        const keywords = words.map(s => s.toLowerCase());
        
        const reply = await db.selectFrom('autoreplyreply')
            .select(['autoreplyreply.id', 'autoreplyreply.reply', 'autoreplyreply.lastUsed'])
            .innerJoin('autoreplyterm', 'autoreplyterm.replyId', 'autoreplyreply.id')
            .where('autoreplyterm.term', 'in', keywords)
            .executeTakeFirst();
        
        if(reply != undefined) {
            const now = Date.now();
            const secondsSinceUse = (now - reply.lastUsed) / 1000
            if(secondsSinceUse > 3600) {
                updateReply(reply.id, now);
                await message.reply(reply.reply)
            }
        }
    });

    client.on(Events.GuildMemberAdd, async (member: GuildMember) => {
        if(config.autoroleEnable) {
            member.roles.add(config.autoroleRole);
        }
    });

    client.on(Events.MessageReactionAdd, async (reaction: MessageReaction | PartialMessageReaction, user: User | PartialUser) => {
        console.log("reaction add");
    })

}