import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from 'discord.js';
import { SlashCommand } from '../../util/slashcommand';
import { createReply, createTerm, db, deleteTerm, getReply, getTerm, updateTerm } from '../../util/db/db';

class AutoReplyCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName('autoreply')
            .setDescription('Create or remove auto-repy messages in this channel.')
            .setDefaultMemberPermissions(8)
            .addSubcommand(subcommand =>
                subcommand
                    .setName('add')
                    .setDescription('Make Elaina reply to a term.')
                    .addStringOption(option => option.setName('term').setDescription('The term to be replied to.').setRequired(true))
                    .addStringOption(option => option.setName('reply').setDescription('The message to reply with.').setRequired(true))
            )
            .addSubcommand(subcommand =>
                subcommand
                    .setName('remove')
                    .setDescription('Make Elaina stop replying to a term.')
                    .addStringOption(option => option.setName('term').setDescription('The term to be removed.').setRequired(true))
            )
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        let term = interaction.options.getString('term');
        if(term == null) // Validate term is present, it should be required but just in case.
            return await interaction.reply({ content: 'Term was not present in command.', ephemeral: true });
        term = term.trim().toLowerCase();

        try {
            if(interaction.options.getSubcommand() == 'add')
                return await this.add(interaction, term);
            if(interaction.options.getSubcommand() == 'remove')
                return await this.remove(interaction, term);
        }
        catch(error) {
            return await interaction.reply({ content: `Elaina can't do this, something went wrong.`, ephemeral: true });
        }

        return await interaction.reply({ content: 'Subcommand not found.', ephemeral: true });
    };



    /**
     * Command handler for adding a reply to a term.
     */
    async add(interaction: ChatInputCommandInteraction, term : string) : Promise<InteractionResponse<boolean>> {
        let reply = interaction.options.getString('reply');
        if(reply == null) // Validate reply is present, it should be required but just in case.
            return await interaction.reply({ content: 'Reply not present in command.', ephemeral: true });
        reply = reply.trim();

        // Grab reply if one with this text already exists and create new one if it doesn't.
        let replyId = (await getReply(reply))?.id;
        if(replyId === undefined)
            replyId = Number((await createReply(reply)).insertId)
        
        const existingTerm = await getTerm(term);
        if(existingTerm === undefined)
            await createTerm(term, replyId);
        else
            await updateTerm(existingTerm.id, replyId);
                
        console.log(`Added term '${term}' with reply '${reply}' to the auto-replies list.`);
        return await interaction.reply({ content: `Elaina will now reply to the term '${term}' with '${reply}'.`, ephemeral: true });
    }



    /**
     * Command handler for removing a reply from a term.
     */
    async remove(interaction: ChatInputCommandInteraction, term : string) : Promise<InteractionResponse<boolean>> {
        const result = await deleteTerm(term);
        if(result.numDeletedRows <= 0)
            interaction.reply({ content: `Elaina could not find a response to '${term}' in her database.`, ephemeral: true });

        console.log(`Removed term '${term}' from the auto-replies list.`);
        return await interaction.reply({ content: `Elaina will stop to replying to '${term}'.`, ephemeral: true });
    }

}

module.exports = new AutoReplyCommand();