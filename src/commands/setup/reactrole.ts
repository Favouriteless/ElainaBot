import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from 'discord.js';
import { SlashCommand } from '../../util/slashcommand';
import { createReactionRole, deleteReactionRole, getReactionRole, updateReactionRole } from '../../util/db/db';

class ReactRoleCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName('reactrole')
            .setDescription('Create or edit reaction role messages in this channel. Only works with custom emojis.')
            .setDefaultMemberPermissions(8)
            .addSubcommand(subcommand =>
                subcommand
                    .setName('add')
                    .setDescription('Add or update a reaction role to a message.')
                    .addStringOption(option => option.setName('message').setDescription('ID of the message to attach a reaction role to.').setRequired(true))
                    .addStringOption(option => option.setName('emoji').setDescription('The emoji to use for reaction role.').setRequired(true))
                    .addRoleOption(option => option.setName('role').setDescription('The role to assign to that emoji.').setRequired(true))
            )
            .addSubcommand(subcommand =>
                subcommand
                    .setName('remove')
                    .setDescription('Remove a reaction role from a message.')
                    .addStringOption(option => option.setName('message').setDescription('ID of the message to attach a reaction role to.').setRequired(true))
                    .addStringOption(option => option.setName('emoji').setDescription('The emoji to use for reaction role.').setRequired(true))
            )   
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        let message = interaction.options.getString('message');
        if(message == null) // Validate term is present, it should be required but just in case.
            return await interaction.reply({ content: 'Message ID was not present in command.', ephemeral: true });
        message = message.trim().toLowerCase();

        if(!/(\d{17,19})/.test(message))
            return await interaction.reply({ content: `Elaina thinks '${message}' is not a valid message ID.`})

            const emojiString = interaction.options.getString('emoji');
        if(emojiString == null) // Validate emoji is present, it should be required but just in case.
            return await interaction.reply({ content: 'emoji was not present in command.', ephemeral: true });
        if(!/<?(a)?:?(\w{2,32}):(\d{17,19})>?/.test(emojiString))
            return await interaction.reply({ content: `Elaina thinks '${emojiString}' is not a valid custom emoji.`})

        try {
            if(interaction.options.getSubcommand() == 'add')
                return await this.add(interaction, message, emojiString);
            if(interaction.options.getSubcommand() == 'remove')
                return await this.remove(interaction, message, emojiString);
        }
        catch(error) {
            console.error(error);
            return await interaction.reply({ content: `Elaina can't do this, something went wrong.`, ephemeral: true });
        }

        return await interaction.reply({ content: 'Subcommand not found.', ephemeral: true });
    };



    /**
     * Command handler for adding a reply to a term.
     */
    async add(interaction: ChatInputCommandInteraction, messageId: string, emoji: string) : Promise<InteractionResponse<boolean>> {
        let role = interaction.options.getRole('role');
        if(role == null) // Validate role is present, it should be required but just in case.
            return await interaction.reply({ content: 'Role was not present in command.', ephemeral: true });        
        const existingEntry = await getReactionRole(messageId, emoji);
        if(existingEntry === undefined)
            await createReactionRole(messageId, emoji, role.id);
        else
            await updateReactionRole(messageId, emoji, role.id);
        
        interaction.channel?.messages.fetch(messageId).then(message => {
            if(message !== undefined)
                message.react(emoji);
        });

        console.log(`Added reaction role '${emoji}' for '${role.id}' on '${messageId}'.`);
        return await interaction.reply({ content: `Elaina will start giving '${role.name}' to '${emoji}' reactions on '${messageId}'.`, ephemeral: true });
    }



    /**
     * Command handler for removing a reply from a term.
     */
    async remove(interaction: ChatInputCommandInteraction, message: string, emoji: string) : Promise<InteractionResponse<boolean>> {
        const result = await deleteReactionRole(message, emoji)
        if(result.numDeletedRows <= 0)
            return await interaction.reply({ content: `Elaina could not find a role for '${emoji}' on '${message}' in her database.`, ephemeral: true });

        interaction.channel?.messages.fetch(message).then(message => {
            const idMatch = emoji.match(/(\d{17,19})/g);
            if(message !== undefined && idMatch !== undefined && idMatch != null)
                message.reactions.cache.get(idMatch[idMatch.length-1])?.remove()
                    .catch(error => console.error(`Failed to remove role reactions from ${message}: ${error}`));
        });
        
        console.log(`Removed reaction role for '${emoji}' on '${message}'.`);
        return await interaction.reply({ content: `Elaina will stop to giving roles for '${emoji}' on '${message}'.`, ephemeral: true });
    }

}

module.exports = new ReactRoleCommand();