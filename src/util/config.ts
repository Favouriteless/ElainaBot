import fs from 'fs';
import path from 'path';

interface Config {
    autoroleRole: string;
    autoroleEnable: boolean;
}


export const config: Config = require('../../data/config.json'); // This path is relative to THIS file.

export function saveConfig() {
    fs.writeFileSync(path.join(__dirname, '../../data/config.json'), JSON.stringify(config, undefined, 1)); // The path here is relative to the "root" file, ie elaina.ts or elaina.js (prod).
}