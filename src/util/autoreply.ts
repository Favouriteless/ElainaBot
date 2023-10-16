import { Events, Message } from "discord.js";
import { Client } from "../elaina";
import { db, updateReply } from "./db/db";

export function registerAutoreplyListeners(client: Client) {

    client.on(Events.MessageCreate, async (message: Message) => {
        if(message.author.bot)
            return;

        let words = message.content.match(/(\?*[\w\-]+)/g); // Regex matches 0/more ? followed by 1/more alphanumeric (or -_) characters
        if(words == null)
            return;
        
        const keywords = words.map(s => s.toLowerCase());
        
        const reply = await db.selectFrom('autoreplyreply')
            .select(['autoreplyreply.id', 'autoreplyreply.reply', 'autoreplyreply.lastUsed'])
            .innerJoin('autoreplyterm', 'autoreplyterm.replyId', 'autoreplyreply.id')
            .where('autoreplyterm.term', 'in', keywords)
            .executeTakeFirst();
        
        if(reply != undefined) {
            const now = Date.now();
            const secondsSinceUse = (now - reply.lastUsed) / 1000
            if(secondsSinceUse > 3600) {
                updateReply(reply.id, now);
                await message.reply(reply.reply)
            }
        }
    });

}
