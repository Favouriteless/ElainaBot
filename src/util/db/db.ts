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

export async function createMacro(macro: string, reply: string) {
	return await db.insertInto('macro')
		.values({ macro: macro, reply: reply });
}

export async function removeMacro(macro: string) {
	return await db.deleteFrom('macro')
		.where('macro.macro', '=', macro);
}

export async function getMacro(macro: string) {
	return await db.selectFrom('macro')
		.selectAll()
		.where('macro.macro', '=', macro)
		.executeTakeFirst();
}