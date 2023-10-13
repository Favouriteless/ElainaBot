import { SlashCommandBuilder, ChatInputCommandInteraction, InteractionResponse } from "discord.js";
import { SlashCommand } from "../util/slashcommand";
import fs from "fs";

interface AutoReplies {
    [key: string]: string
}

class AutoReplyCommand extends SlashCommand {

    replies: AutoReplies = {};

    constructor() {
        super(new SlashCommandBuilder()
            .setName("autoreply")
            .setDescription("Create or remove auto-repy messages in this channel.")
            .setDefaultMemberPermissions(8)
            .addSubcommand(subcommand =>
                subcommand
                    .setName("add")
                    .setDescription("Add or override an autoreply term.")
                    .addStringOption(option => option.setName("term").setDescription("The term to be replied to.").setRequired(true))
                    .addStringOption(option => option.setName("reply").setDescription("The message to reply with.").setRequired(true))
            )
            .addSubcommand(subcommand =>
                subcommand
                    .setName("remove")
                    .setDescription("Remove an autoreply term.")
                    .addStringOption(option => option.setName("term").setDescription("The term to be removed.").setRequired(true))
            )
        );
        this.replies = require('../data/autoreply-terms.json');
    }

    async execute(interaction: ChatInputCommandInteraction) : Promise<InteractionResponse<boolean>> {
        let term = interaction.options.getString("term");
        if(term == null) // Validate term is present, it should be required but just in case.
            return await interaction.reply({ content: "Term was not found.", ephemeral: true });

        term = term.toLowerCase();

        if(interaction.options.getSubcommand() == "add")
            return await this.add(interaction, term);
        if(interaction.options.getSubcommand() == "remove")
            return await this.remove(interaction, term);

        return await interaction.reply({ content: "Subcommand not found.", ephemeral: true });
    };

    async add(interaction: ChatInputCommandInteraction, term : string) : Promise<InteractionResponse<boolean>> {
        const reply = interaction.options.getString("reply");
        if(reply == null) // Validate reply is present, it should be required but just in case.
            return await interaction.reply({ content: "Reply was not found.", ephemeral: true });

        this.replies[term] = reply;
        if(!this.saveReplies()) 
            return await interaction.reply({ content: `Something went wrong. Elaina can't reply to '${term}' with '${reply}'.`, ephemeral: true });

        console.log(`Added term '${term}' with reply '${reply}' to the auto-replies list.`);
        return await interaction.reply({ content: `Elaina will now reply to the term '${term}' with '${reply}'.`, ephemeral: true });
    }

    async remove(interaction: ChatInputCommandInteraction, term : string) : Promise<InteractionResponse<boolean>> {
        delete this.replies[term];
        if(!this.saveReplies()) 
            return await interaction.reply({ content: `Something went wrong. Elaina will continue to reply to '${term}'.`, ephemeral: true });

        return await interaction.reply({ content: `Elaina will stop to replying to '${term}'.`, ephemeral: true });
    }

    saveReplies() : boolean {
        try {
            fs.writeFileSync("data/autoreply-terms.json", JSON.stringify(this.replies));
        } catch (err) {
            console.error(err);
            return false;
        }
        return true;
    }

}

module.exports = new AutoReplyCommand();