import { ChatInputCommandInteraction, InteractionResponse, SlashCommandBuilder } from "discord.js";

export abstract class SlashCommand {
    data: SlashCommandBuilder;

    constructor(data: SlashCommandBuilder) {
        this.data = data;
    }

    abstract execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>>;
}