import { Events, BaseInteraction, Message } from "discord.js";
import { Client } from "..";
import { db } from "./db/db";

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

        const words = message.content.split(/([_\W])/).map(s => s.toLowerCase());
        
        const reply = await db.selectFrom('autoreplyreply')
            .select('autoreplyreply.reply')
            .innerJoin('autoreplyterm', 'autoreplyterm.replyId', 'autoreplyreply.id')
            .where('autoreplyterm.term', 'in', words)
            .executeTakeFirst();
        
        if(reply != undefined) {
            await message.reply(reply.reply)
        }

    });

}
