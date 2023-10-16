import { Events, BaseInteraction, Message, GuildMember, MessageReaction, User, PartialMessageReaction, PartialUser } from "discord.js";
import { db, getReactionRole, updateReply } from "./db/db";
import { Client } from "../elaina";
import { config } from "./config";
import { registerReactionRoleListeners } from "./reaction-role";
import { registerAutoreplyListeners } from "./autoreply";

export function registerListeners(client: Client) {

    client.once(Events.ClientReady, c => console.log(`Ready! Logged in as ${c.user.tag}`));
    client.on(Events.GuildMemberAdd, async (member: GuildMember) => { if(config.autoroleEnable) member.roles.add(config.autoroleRole); });

    // Slash command listener
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

    registerAutoreplyListeners(client);
    registerReactionRoleListeners(client);

}