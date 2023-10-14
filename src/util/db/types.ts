import { Generated, Insertable, Selectable, Updateable } from "kysely";

// ----------------------------- DATABASE -----------------------------

export interface Database {
	autoreplyreply: AutoreplyReplyTable;
	autoreplyterm: AutoreplyTermTable;
}

// ------------------------------ TABLES ------------------------------

export interface AutoreplyReplyTable {
	id: Generated<number>;
	reply: string;
	lastUsed: Generated<number>;
}
export type AutoreplyReply = Selectable<AutoreplyReplyTable>;
export type NewAutoreplyReply = Insertable<AutoreplyReplyTable>;
export type AutoreplyReplyUpdate = Updateable<AutoreplyReplyTable>;


export interface AutoreplyTermTable {
	id: Generated<number>;
	term: string;
	replyId: number;
}
export type AutoreplyTerm = Selectable<AutoreplyTermTable>;
export type NewAutoreplyTerm = Insertable<AutoreplyTermTable>;
export type AutoreplyTermUpdate = Updateable<AutoreplyTermTable>;