import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from "discord.js";
import { SlashCommand } from "../util/slashcommand";
  
class ReactRoleCommand extends SlashCommand {

    constructor() {
        super(new SlashCommandBuilder()
            .setName("reactrole")
            .setDescription("Create or edit reaction role messages in this channel.")
            .setDefaultMemberPermissions(8)
            // .addSubcommand(subcommand => 
            //     subcommand
            //         .setName("add")
            // )
        );
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        return await interaction.reply({ content: "Role reaction created.", ephemeral: true });
    };
}

module.exports = new ReactRoleCommand();