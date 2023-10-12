import { SlashCommandStringOption, BaseInteraction, ChatInputCommandInteraction, InteractionResponse } from "discord.js";
import { SlashCommand } from "../slashcommand";

const { SlashCommandBuilder } = require('discord.js')
  
class SetRoleMessageCommand extends SlashCommand {
    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        return await interaction.reply({ content: "Message ID Set.", ephemeral: true });
    };
}



module.exports = new SetRoleMessageCommand(new SlashCommandBuilder()
    .setName("setrolemessage")
    .setDescription("Sets the message ID used to assign reaction roles to users.")
    .addStringOption((option: SlashCommandStringOption) =>
        option.setName("message_id")
        .setDescription("ID of the reaction roles message.")
        .setRequired(true)
        ));