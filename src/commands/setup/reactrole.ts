import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from 'discord.js';
import { SlashCommand } from '../../util/slashcommand';
import { createReactionRole, deleteReactionRole, getReactionRole, updateReactionRole } from '../../util/db/db';
  
class ReactRoleCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName('reactrole')
            .setDescription('Create or edit reaction role messages in this channel.')
            .setDefaultMemberPermissions(8)
            .addSubcommand(subcommand =>
                subcommand
                    .setName('add')
                    .setDescription('Add or update a reaction role on a message.')
                    .addStringOption(option => option.setName('message').setDescription('ID of the message to attach a reaction role to.').setRequired(true))
                    .addStringOption(option => option.setName('emote').setDescription('ID of the emote to use for reaction role.').setRequired(true))
                    .addRoleOption(option => option.setName('role').setDescription('ID of the role to assign to that emote.').setRequired(true))
            )
            .addSubcommand(subcommand =>
                subcommand
                    .setName('remove')
                    .setDescription('Remove a reaction role from a message.')
                    .addStringOption(option => option.setName('message').setDescription('ID of the message to attach a reaction role to.').setRequired(true))
                    .addStringOption(option => option.setName('emote').setDescription('ID of the emote to use for reaction role.').setRequired(true))
            )   
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        let message = interaction.options.getString('message');
        if(message == null) // Validate term is present, it should be required but just in case.
            return await interaction.reply({ content: 'Message ID was not present in command.', ephemeral: true });
        message = message.trim().toLowerCase();

        if(/([0-9]+)/.test(message))
            return await interaction.reply({ content: `Elaina thinks '${message}' is not a valid message ID.`})

        let emoteId = interaction.options.getString('emote');
        if(emoteId == null) // Validate emote is present, it should be required but just in case.
            return await interaction.reply({ content: 'Emote was not present in command.', ephemeral: true });

        try {
            if(interaction.options.getSubcommand() == 'emote')
                return await this.add(interaction, message, emoteId);
            if(interaction.options.getSubcommand() == 'remove')
                return await this.remove(interaction, message, emoteId);
        }
        catch(error) {
            return await interaction.reply({ content: `Elaina can't do this, something went wrong.`, ephemeral: true });
        }

        return await interaction.reply({ content: 'Subcommand not found.', ephemeral: true });
    };



    /**
     * Command handler for adding a reply to a term.
     */
    async add(interaction: ChatInputCommandInteraction, message: string, emote: string) : Promise<InteractionResponse<boolean>> {
        let role = interaction.options.getRole('role');
        if(role == null) // Validate role i spresent, it should be required but just in case.
            return await interaction.reply({ content: 'Role was not present in command.', ephemeral: true });        
        const existingEntry = await getReactionRole(message, emote);
        if(existingEntry === null)
            await createReactionRole(message, emote, role.id);
        else
            await updateReactionRole(message, emote, role.id);

        console.log(`Added reaction role '${emote}' for '${role.id}' on '${message}'.`);
        return await interaction.reply({ content: `Elaina will start giving '${role.id} to '${emote}' reactions on '${message}'.`, ephemeral: true });
    }



    /**
     * Command handler for removing a reply from a term.
     */
    async remove(interaction: ChatInputCommandInteraction, message: string, emote: string) : Promise<InteractionResponse<boolean>> {
        const result = await deleteReactionRole(message, emote);
        if(result.numDeletedRows <= 0)
            interaction.reply({ content: `Elaina could not find a role for '${emote}' on '${message}' in her database.`, ephemeral: true });

        console.log(`Removed reaction role for '${emote}' on '${message}'.`);
        return await interaction.reply({ content: `Elaina will stop to giving roles for '${emote}' on '${message}'.`, ephemeral: true });
    }

}

module.exports = new ReactRoleCommand();