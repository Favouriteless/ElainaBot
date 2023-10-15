import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from 'discord.js';
import { SlashCommand } from '../../util/slashcommand';
import { config, saveConfig } from '../../util/config';

class AutoRoleCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName('autorole')
            .setDescription('Enable, disable or edit the default role for users.')
            .setDefaultMemberPermissions(8)
            .addSubcommand(subcommand =>
                subcommand
                    .setName('enable')
                    .setDescription('Make Elaina give a role to new users.')
                    .addRoleOption(option => option.setName('role').setDescription('The role to be assigned.').setRequired(true))
            )
            .addSubcommand(subcommand =>
                subcommand
                    .setName('disable')
                    .setDescription('Make Elaina stop giving roles to new users.')
            )
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        if(interaction.options.getSubcommand() == 'enable')
            return await this.enable(interaction);
        if(interaction.options.getSubcommand() == 'disable')
            return await this.disable(interaction);

        return await interaction.reply({ content: 'Subcommand not found.', ephemeral: true });
    };



    /**
     * Command handler for enabling/setting the auto-role role.
     */
    async enable(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        let role = interaction.options.getRole('role');
        if(role == null) // Validate reply is present, it should be required but just in case.
            return await interaction.reply({ content: 'Reply not present in command.', ephemeral: true });

        config.autoroleRole = role.id;
        config.autoroleEnable = true;
        saveConfig();
                
        return await interaction.reply({ content: `Elaina will now give the ${role.name} role to new members.`, ephemeral: true });
    }



    /**
     * Command handler for disabling auto-role.
     */
    async disable(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        config.autoroleEnable = false;
        saveConfig();
        return await interaction.reply({ content: `Elaina will stop giving roles to new members.`, ephemeral: true });
    }

}

module.exports = new AutoRoleCommand();