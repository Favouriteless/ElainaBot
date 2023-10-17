import { Collection } from 'discord.js';
import { SlashCommand } from './slashcommand';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Populates commands collection with all slash commands from the commands directory and it's subfolders. Not particularly fast, but only needs to run once on startup.
 * @param dir The path of the commands directory.
 * @param commands (Optional) {@link Collection<string, SlashCommand>} containing any existing commands. If not provided, an empty collection will be used.
 * @returns A collection of all {@link SlashCommand}s and their names, mapped name -> command.
 */
export function loadCommands(dir: string, commands: Collection<string, SlashCommand> = new Collection()) : Collection<string, SlashCommand> {
    let allPaths = fs.readdirSync(dir);

    const commandFiles = allPaths.filter((_path: string) => _path.endsWith('.ts') || _path.endsWith('.js')) // All files ending with .ts or js are asumed to be commands which need loading.
    const subDirs = allPaths.filter((_path: string) => !_path.includes('.')) // Paths which don't contain "." are assumed to be a subdirectory.

    for(const i in subDirs) {
        loadCommands(path.join(dir, subDirs[i]), commands); // Load all subdirs -- these are checked recursively.
    }
    
    for (const file of commandFiles) { // Load the command.ts files in this directory.
        const filePath = path.join(dir, file);
        const command = require(filePath);

        if (command instanceof SlashCommand)
            commands.set(command.data.name, command);
        else
            console.log(`[WARNING] The command file at ${filePath} does not export a valid SlashCommand.`);
    }

    return commands;
}