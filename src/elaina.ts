import { loadCommands } from './util/command-loader';
import { Client as DJSClient, GatewayIntentBits, Collection, Partials } from 'discord.js';
import { registerListeners } from './util/event-listeners';
import { SlashCommand } from './util/slashcommand';
import path from 'node:path';
const { token } = require('../data/bot-details.json');


export class Client extends DJSClient  {
    commands: Collection<string, SlashCommand> = loadCommands(path.join(__dirname, './commands')); // Populate client commands list.
}

const client = new Client({ intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.GuildMessageReactions,
    GatewayIntentBits.MessageContent,
    GatewayIntentBits.GuildMembers
], partials: [Partials.Reaction, Partials.Message, Partials.User] });
registerListeners(client);
client.login(token);