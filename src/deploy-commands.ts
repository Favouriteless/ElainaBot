import { REST, Routes } from 'discord.js';
import { loadCommands } from './util/command-loader';
import path from 'path';
const { clientId, token } = require('../data/bot-details.json');

const commandJsons = [];
const commandFiles = loadCommands(path.join(__dirname, './commands'));

for (const command of commandFiles) {
	commandJsons.push(command[1].data.toJSON()); // Construct JSON objects from the commands in the folder
}

const rest = new REST().setToken(token);

(async () => {
	try {
		console.log(`Started refreshing application (/) commands.`);

		// The put method is used to fully refresh all commands in the guild with the current set
		await rest.put(Routes.applicationCommands(clientId), { body: commandJsons });
		console.log(`Successfully reloaded application (/) commands.`);
	} catch (error) {
		console.error(error);
	}
})();