import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from 'discord.js';
import { SlashCommand } from '../../util/slashcommand';
import { config, saveConfig } from '../../util/config';

class HelloEmojiCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName('macro')
            .setDescription('Set the emoji Elaina uses to say hi.')
            .setDefaultMemberPermissions(8)
            .addStringOption(option => option.setName('emoji').setDescription('Emoji for Elaina to say hi with.').setRequired(true))
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        const emojiString = interaction.options.getString('emoji');
        if(emojiString == null) // Validate emoji is present, it should be required but just in case.
            return await interaction.reply({ content: 'emoji was not present in command.', ephemeral: true });
        if(!/<?(a)?:?(\w{2,32}):(\d{17,19})>?/.test(emojiString))
            return await interaction.reply({ content: `Elaina thinks '${emojiString}' is not a valid custom emoji.`})

        config.helloEmoji = emojiString;
        config.helloEmojiEnabled = true;
        saveConfig();
                
        return await interaction.reply({ content: `Elaina will now say hi with '${emojiString} when mentioned.`, ephemeral: true });
    };
}

module.exports = new HelloEmojiCommand();