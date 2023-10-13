import { loadCommands } from './util/command-loader';
import { Client as DJSClient, GatewayIntentBits, Collection, ClientOptions } from 'discord.js';
import { registerListeners } from './util/event-listeners';
import { SlashCommand } from './util/slashcommand';
const path = require('node:path');
const { token } = require('../config.json');

export class Client extends DJSClient  {
    commands: Collection<string, SlashCommand> = new Collection();
}

(async () => { // Run this in an anonymous self-executing async function as loadCommands is also async.
    const client = new Client({ intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.GuildMessageReactions, GatewayIntentBits.MessageContent] });
    await loadCommands(path.join(__dirname, './commands'), client.commands); // Populate client commands list.
    registerListeners(client);
    client.login(token);
})();




