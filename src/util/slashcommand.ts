import { ChatInputCommandInteraction, InteractionResponse, SlashCommandBuilder, SlashCommandSubcommandBuilder, SlashCommandSubcommandsOnlyBuilder } from "discord.js";

export type CommandBuilder = SlashCommandBuilder | SlashCommandSubcommandBuilder | SlashCommandSubcommandsOnlyBuilder;

export abstract class SlashCommand {
    data: CommandBuilder;

    constructor(data: CommandBuilder) {
        this.data = data;
    }

    abstract execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>>;
}