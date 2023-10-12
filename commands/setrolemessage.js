const { SlashCommandBuilder } = require('discord.js')

module.exports = {
    
    data: new SlashCommandBuilder()
        .setName("setrolemessage")
        .setDescription("Sets the message ID used to assign reaction roles to users.")
        .addStringOption(option =>
            option.setName("message_id")
            .setDescription("ID of the reaction roles message.")
            .setRequired(true)
            ),
        
    async execute(interaction) {
        await interaction.reply("Message ID set.");
    }

};