import { loadCommands } from './util/command-loader';
import { Client as DJSClient, GatewayIntentBits, Collection } from 'discord.js';
import { registerListeners } from './util/event-listeners';
import { SlashCommand } from './util/slashcommand';
const path = require('node:path');
const { token } = require('./data/config.json');

export class Client extends DJSClient  {
    commands: Collection<string, SlashCommand> = loadCommands(path.join(__dirname, './commands')); // Populate client commands list.
}

const client = new Client({ intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.GuildMessageReactions, GatewayIntentBits.MessageContent] });
registerListeners(client);
client.login(token);

