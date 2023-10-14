import { Kysely, SqliteDialect } from 'kysely';
import { Database } from './types';
import SQLite from 'better-sqlite3';
import path from 'path';

const dialect = new SqliteDialect({
	database: new SQLite(path.join(__dirname, '../../../data/database.db'))
});

export const db = new Kysely<Database>({
	dialect
});



export async function getTerm(term: string) {
	return await db.selectFrom('autoreplyterm')
                .selectAll()
                .where('autoreplyterm.term', '=', term)
                .executeTakeFirst();
}

export async function createTerm(term: string, replyId: number) {
    return await db.insertInto('autoreplyterm')
        .values({ term: term, replyId: replyId })
        .executeTakeFirst();
}

export async function updateTerm(id: number, replyId: number) {
    return await db.updateTable('autoreplyterm')
        .set({ replyId: replyId })
        .where('autoreplyterm.id', '=', id)
        .executeTakeFirst();
}

export async function deleteTerm(term: string) {
        return await db.deleteFrom('autoreplyterm')
                .where('autoreplyterm.term', '=', term)
                .executeTakeFirst()
}

export async function getReply(reply: string) {
	return await db.selectFrom('autoreplyreply')
                .selectAll()
                .where('autoreplyreply.reply', '=', reply)
                .executeTakeFirst();
}

export async function updateReply(id: number, lastUsed: number) {
        return await db.updateTable('autoreplyreply')
                .set({ lastUsed: lastUsed })
                .where('autoreplyreply.id', '=', id)
                .executeTakeFirst();
}

export async function createReply(reply: string) {
    return await db.insertInto('autoreplyreply')
        .values({ reply: reply })
        .executeTakeFirst();
}