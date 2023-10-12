import { Collection } from 'discord.js';
import { SlashCommand } from './slashcommand';
const fs = require('node:fs');
const path = require('node:path');

export function loadCommands() : Collection<string, SlashCommand> {
    const commands = new Collection<string, SlashCommand>();
    const commandsPath = path.join(__dirname, 'commands');
    const commandFiles = fs.readdirSync(commandsPath).filter((file: string) => file.endsWith('.ts'))
    
    for (const file of commandFiles) {
        const filePath = path.join(commandsPath, file);
        const command = require(filePath);

        if ('data' in command && 'execute' in command)
            commands.set(command.data.name, command);
        else
            console.log(`[WARNING] The command at ${filePath} is missing a required "data" or "execute" property.`);
    }
    return commands;
}