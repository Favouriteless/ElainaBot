import { loadCommands } from './command-loader';
import { Client as DJSClient, GatewayIntentBits, Collection } from 'discord.js';
import { registerListeners } from './event-listeners';
const { token } = require('../config.json');

export class Client extends DJSClient {
    commands = loadCommands();
}

const client = new Client({ intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.GuildMessageReactions, GatewayIntentBits.MessageContent] });
registerListeners(client);
client.login(token);