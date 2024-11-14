import fs from 'fs';
import path from 'path';
import configFile from '../../data/config.json'

interface Config {
    helloEmoji: string;
    helloEmojiEnabled: boolean;
}


export const config: Config = configFile; // This path is relative to THIS file.

export function saveConfig() {
    fs.writeFileSync(path.join(__dirname, '../../data/config.json'), JSON.stringify(config, undefined, 1)); // The path here is relative to the "root" file, ie elaina.ts or elaina.js (prod).
}